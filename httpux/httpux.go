package httpux

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// ClientInterface is the general interface provided to perform HTTP-over-UDP
// requests.
type ClientInterface interface {
	// Do performs a request. The timeout is how long to wait for before returning
	// the responses that were received. An error is only returned for failing to
	// send the request. Failures in receipt simply do not add to the resulting
	// responses.
	Do(
		req *http.Request,
		interval time.Duration,
	) error

	ReceiveChan() chan *http.Response
}

type ClientInterfaceCtx interface {
	// DoWithContext performs a request. If the input request has a
	// deadline, then that value will be used as the timeout for how long
	// to wait before returning the responses that were received. If the
	// request's context is canceled, this method will return immediately.
	//
	// If the request's context is never canceled, and does not have a
	// deadline, then this function WILL NEVER RETURN. You MUST set an
	// appropriate deadline on the context, or otherwise cancel it when you
	// want to finish an operation.
	//
	// An error is only returned for failing to send the request. Failures
	// in receipt simply do not add to the resulting responses.
	DoWithContext(
		req *http.Request,
		interval time.Duration,
	) error

	ReceiveChan() chan *http.Response
}

// HTTPUClient is a client for dealing with HTTPU (HTTP over UDP). Its typical
// function is for HTTPMU, and particularly SSDP.
type HTTPUClient struct {
	connLock sync.Mutex // Protects use of conn.
	conn     net.PacketConn
	receiver chan *http.Response
}

// NewHTTPUClient creates a new HTTPUClient, opening up a new UDP socket for the
// purpose.
func NewHTTPUClient() (*HTTPUClient, error) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, err
	}
	return &HTTPUClient{conn: conn, receiver: make(chan *http.Response)}, nil
}

// NewHTTPUClientAddr creates a new HTTPUClient which will broadcast packets
// from the specified address, opening up a new UDP socket for the purpose
func NewHTTPUClientAddr(addr string) (*HTTPUClient, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.New("Invalid listening address")
	}
	conn, err := net.ListenPacket("udp", ip.String()+":0")
	if err != nil {
		return nil, err
	}
	return &HTTPUClient{conn: conn}, nil
}

// Close shuts down the client. The client will no longer be useful following
// this.
func (httpu *HTTPUClient) Close() error {
	httpu.connLock.Lock()
	defer httpu.connLock.Unlock()
	close(httpu.receiver)
	return httpu.conn.Close()
}

// Do implements ClientInterface.Do.
//
// Note that at present only one concurrent connection will happen per
// HTTPUClient.
func (httpu *HTTPUClient) Do(
	req *http.Request,
	interval time.Duration,
) error {
	return httpu.DoWithContext(req, interval)
}

// DoWithContext implements ClientInterfaceCtx.DoWithContext.
//
// Make sure to read the documentation on the ClientInterfaceCtx interface
// regarding cancellation!
func (httpu *HTTPUClient) DoWithContext(
	req *http.Request,
	interval time.Duration,
) error {
	tasks := &errgroup.Group{}
	tasks.Go(func() error {
		return httpu.startLoopSendHTTPURequest(req, interval)
	})

	tasks.Go(func() error {
		return httpu.startReceiveResponse(req)
	})
	return tasks.Wait()
}

func (httpu *HTTPUClient) startLoopSendHTTPURequest(req *http.Request, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	ctx := req.Context()
	for {
		select {
		case <-ticker.C:
			err := httpu.sendHTTPURequest(req)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
	return nil
}

func (httpu *HTTPUClient) sendHTTPURequest(req *http.Request) error {
	// Create the request. This is a subset of what http.Request.Write does
	// deliberately to avoid creating extra fields which may confuse some
	// devices.
	var requestBuf bytes.Buffer
	method := req.Method
	if method == "" {
		method = "GET"
	}
	if _, err := fmt.Fprintf(&requestBuf, "%s %s HTTP/1.1\r\n", method, req.URL.RequestURI()); err != nil {
		return err
	}
	if err := req.Header.Write(&requestBuf); err != nil {
		return err
	}
	if _, err := requestBuf.Write([]byte{'\r', '\n'}); err != nil {
		return err
	}

	destAddr, err := net.ResolveUDPAddr("udp", req.Host)
	if err != nil {
		return err
	}

	ctx := req.Context()
	// Handle context cancelation
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			// if context is cancelled, stop any connections by setting time in the past.
			httpu.conn.SetDeadline(time.Now().Add(-time.Second))
		case <-done:
		}
	}()

	// Send request.
	if n, err := httpu.conn.WriteTo(requestBuf.Bytes(), destAddr); err != nil {
		return err
	} else if n < len(requestBuf.Bytes()) {
		return fmt.Errorf("httpu: wrote %d bytes rather than full %d in request",
			n, len(requestBuf.Bytes()))
	}
	return nil
}

func (httpu *HTTPUClient) startReceiveResponse(req *http.Request) error {
	responseBytes := make([]byte, 2048)
	for {
		// 2048 bytes should be sufficient for most networks.
		n, _, err := httpu.conn.ReadFrom(responseBytes)
		if err != nil {
			if err, ok := err.(net.Error); ok {
				if err.Timeout() {
					break
				}
				if err.Temporary() {
					// Sleep in case this is a persistent error to avoid pegging CPU until deadline.
					time.Sleep(10 * time.Millisecond)
					continue
				}
			}
			return err
		}

		// Parse response.
		response, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(responseBytes[:n])), req)
		if err != nil {
			log.Printf("httpu: error while parsing response: %v", err)
			continue
		} else { //reik
			//fmt.Println("response ", response)
		}

		// Set the related local address used to discover the device.
		if a, ok := httpu.conn.LocalAddr().(*net.UDPAddr); ok {
			response.Header.Add(LocalAddressHeader, a.IP.String())
		}
	}
	return nil
}

const LocalAddressHeader = "goupnp-local-address"
