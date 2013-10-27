package soap

import (
	"bytes"
	"math"
	"testing"
	"time"
)

type convTest interface {
	Marshal() (string, error)
	Unmarshal(string) (interface{}, error)
	Equal(result interface{}) bool
}

// duper is an interface that convTest values may optionally also implement to
// generate another convTest for a value in an otherwise identical testCase.
type duper interface {
	Dupe(tag string) convTest
}

type testCase struct {
	value            convTest
	str              string
	wantMarshalErr   bool
	wantUnmarshalErr bool
	noMarshal        bool
	noUnMarshal      bool
	tag              string
}

type Fixed14_4Test float64

func (v Fixed14_4Test) Marshal() (string, error) {
	return MarshalFixed14_4(float64(v))
}
func (v Fixed14_4Test) Unmarshal(s string) (interface{}, error) {
	return UnmarshalFixed14_4(s)
}
func (v Fixed14_4Test) Equal(result interface{}) bool {
	return math.Abs(float64(v)-result.(float64)) < 0.001
}

type CharTest rune

func (v CharTest) Marshal() (string, error) {
	return MarshalChar(rune(v))
}
func (v CharTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalChar(s)
}
func (v CharTest) Equal(result interface{}) bool {
	return rune(v) == result.(rune)
}

type DateTest struct{ time.Time }

func (v DateTest) Marshal() (string, error) {
	return MarshalDate(time.Time(v.Time))
}
func (v DateTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalDate(s)
}
func (v DateTest) Equal(result interface{}) bool {
	return v.Time.Equal(result.(time.Time))
}
func (v DateTest) Dupe(tag string) convTest {
	if tag != "no:dateTime" {
		return DateTimeTest{v.Time}
	}
	return nil
}

type TimeOfDayTest struct {
	TimeOfDay
}

func (v TimeOfDayTest) Marshal() (string, error) {
	return MarshalTimeOfDay(v.TimeOfDay)
}
func (v TimeOfDayTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalTimeOfDay(s)
}
func (v TimeOfDayTest) Equal(result interface{}) bool {
	return v.TimeOfDay == result.(TimeOfDay)
}
func (v TimeOfDayTest) Dupe(tag string) convTest {
	if tag != "no:time.tz" {
		return TimeOfDayTzTest{v.TimeOfDay}
	}
	return nil
}

type TimeOfDayTzTest struct {
	TimeOfDay
}

func (v TimeOfDayTzTest) Marshal() (string, error) {
	return MarshalTimeOfDayTz(v.TimeOfDay)
}
func (v TimeOfDayTzTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalTimeOfDayTz(s)
}
func (v TimeOfDayTzTest) Equal(result interface{}) bool {
	return v.TimeOfDay == result.(TimeOfDay)
}

type DateTimeTest struct{ time.Time }

func (v DateTimeTest) Marshal() (string, error) {
	return MarshalDateTime(time.Time(v.Time))
}
func (v DateTimeTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalDateTime(s)
}
func (v DateTimeTest) Equal(result interface{}) bool {
	return v.Time.Equal(result.(time.Time))
}
func (v DateTimeTest) Dupe(tag string) convTest {
	if tag != "no:dateTime.tz" {
		return DateTimeTzTest{v.Time}
	}
	return nil
}

type DateTimeTzTest struct{ time.Time }

func (v DateTimeTzTest) Marshal() (string, error) {
	return MarshalDateTimeTz(time.Time(v.Time))
}
func (v DateTimeTzTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalDateTimeTz(s)
}
func (v DateTimeTzTest) Equal(result interface{}) bool {
	return v.Time.Equal(result.(time.Time))
}

type BooleanTest bool

func (v BooleanTest) Marshal() (string, error) {
	return MarshalBoolean(bool(v))
}
func (v BooleanTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalBoolean(s)
}
func (v BooleanTest) Equal(result interface{}) bool {
	return bool(v) == result.(bool)
}

type BinBase64Test []byte

func (v BinBase64Test) Marshal() (string, error) {
	return MarshalBinBase64([]byte(v))
}
func (v BinBase64Test) Unmarshal(s string) (interface{}, error) {
	return UnmarshalBinBase64(s)
}
func (v BinBase64Test) Equal(result interface{}) bool {
	return bytes.Equal([]byte(v), result.([]byte))
}

type BinHexTest []byte

func (v BinHexTest) Marshal() (string, error) {
	return MarshalBinHex([]byte(v))
}
func (v BinHexTest) Unmarshal(s string) (interface{}, error) {
	return UnmarshalBinHex(s)
}
func (v BinHexTest) Equal(result interface{}) bool {
	return bytes.Equal([]byte(v), result.([]byte))
}

func Test(t *testing.T) {
	const time010203 time.Duration = (1*3600 + 2*60 + 3) * time.Second
	const time0102 time.Duration = (1*3600 + 2*60) * time.Second
	const time01 time.Duration = (1 * 3600) * time.Second
	const time235959 time.Duration = (23*3600 + 59*60 + 59) * time.Second

	// Fake out the local time for the implementation.
	localLoc = time.FixedZone("Fake/Local", 6*3600)
	defer func() {
		localLoc = time.Local
	}()

	tests := []testCase{
		// fixed.14.4
		{str: "0.0000", value: Fixed14_4Test(0)},
		{str: "1.0000", value: Fixed14_4Test(1)},
		{str: "1.2346", value: Fixed14_4Test(1.23456)},
		{str: "-1.0000", value: Fixed14_4Test(-1)},
		{str: "-1.2346", value: Fixed14_4Test(-1.23456)},
		{str: "10000000000000.0000", value: Fixed14_4Test(1e13)},
		{str: "100000000000000.0000", value: Fixed14_4Test(1e14), wantMarshalErr: true, wantUnmarshalErr: true},
		{str: "-10000000000000.0000", value: Fixed14_4Test(-1e13)},
		{str: "-100000000000000.0000", value: Fixed14_4Test(-1e14), wantMarshalErr: true, wantUnmarshalErr: true},

		// char
		{str: "a", value: CharTest('a')},
		{str: "z", value: CharTest('z')},
		{str: "\u1234", value: CharTest(0x1234)},
		{str: "aa", value: CharTest(0), wantMarshalErr: true, wantUnmarshalErr: true},
		{str: "", value: CharTest(0), wantMarshalErr: true, wantUnmarshalErr: true},

		// date
		{str: "2013-10-08", value: DateTest{time.Date(2013, 10, 8, 0, 0, 0, 0, localLoc)}, tag: "no:dateTime"},
		{str: "20131008", value: DateTest{time.Date(2013, 10, 8, 0, 0, 0, 0, localLoc)}, noMarshal: true, tag: "no:dateTime"},
		{str: "2013-10-08T10:30:50", value: DateTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime"},
		{str: "2013-10-08T10:30:50Z", value: DateTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime"},
		{str: "", value: DateTest{}, wantMarshalErr: true, wantUnmarshalErr: true, noMarshal: true},
		{str: "-1", value: DateTest{}, wantUnmarshalErr: true, noMarshal: true},

		// time
		{str: "00:00:00", value: TimeOfDayTest{TimeOfDay{FromMidnight: 0}}},
		{str: "000000", value: TimeOfDayTest{TimeOfDay{FromMidnight: 0}}, noMarshal: true},
		{str: "24:00:00", value: TimeOfDayTest{TimeOfDay{FromMidnight: 24 * time.Hour}}, noMarshal: true}, // ISO8601 special case
		{str: "24:01:00", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "24:00:01", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "25:00:00", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "00:60:00", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "00:00:60", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "01:02:03", value: TimeOfDayTest{TimeOfDay{FromMidnight: time010203}}},
		{str: "010203", value: TimeOfDayTest{TimeOfDay{FromMidnight: time010203}}, noMarshal: true},
		{str: "23:59:59", value: TimeOfDayTest{TimeOfDay{FromMidnight: time235959}}},
		{str: "235959", value: TimeOfDayTest{TimeOfDay{FromMidnight: time235959}}, noMarshal: true},
		{str: "01:02", value: TimeOfDayTest{TimeOfDay{FromMidnight: time0102}}, noMarshal: true},
		{str: "0102", value: TimeOfDayTest{TimeOfDay{FromMidnight: time0102}}, noMarshal: true},
		{str: "01", value: TimeOfDayTest{TimeOfDay{FromMidnight: time01}}, noMarshal: true},
		{str: "foo 01:02:03", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "foo\n01:02:03", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03 foo", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03\nfoo", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03Z", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03+01", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03+01:23", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03+0123", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03-01", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03-01:23", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},
		{str: "01:02:03-0123", value: TimeOfDayTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:time.tz"},

		// time.tz
		{str: "24:00:01", value: TimeOfDayTzTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "01Z", value: TimeOfDayTzTest{TimeOfDay{time01, true, 0}}, noMarshal: true},
		{str: "01:02:03Z", value: TimeOfDayTzTest{TimeOfDay{time010203, true, 0}}},
		{str: "01+01", value: TimeOfDayTzTest{TimeOfDay{time01, true, 3600}}, noMarshal: true},
		{str: "01:02:03+01", value: TimeOfDayTzTest{TimeOfDay{time010203, true, 3600}}, noMarshal: true},
		{str: "01:02:03+01:23", value: TimeOfDayTzTest{TimeOfDay{time010203, true, 3600 + 23*60}}},
		{str: "01:02:03+0123", value: TimeOfDayTzTest{TimeOfDay{time010203, true, 3600 + 23*60}}, noMarshal: true},
		{str: "01:02:03-01", value: TimeOfDayTzTest{TimeOfDay{time010203, true, -3600}}, noMarshal: true},
		{str: "01:02:03-01:23", value: TimeOfDayTzTest{TimeOfDay{time010203, true, -(3600 + 23*60)}}},
		{str: "01:02:03-0123", value: TimeOfDayTzTest{TimeOfDay{time010203, true, -(3600 + 23*60)}}, noMarshal: true},

		// dateTime
		{str: "2013-10-08T00:00:00", value: DateTimeTest{time.Date(2013, 10, 8, 0, 0, 0, 0, localLoc)}, tag: "no:dateTime.tz"},
		{str: "20131008", value: DateTimeTest{time.Date(2013, 10, 8, 0, 0, 0, 0, localLoc)}, noMarshal: true},
		{str: "2013-10-08T10:30:50", value: DateTimeTest{time.Date(2013, 10, 8, 10, 30, 50, 0, localLoc)}, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50T", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true},
		{str: "2013-10-08T10:30:50+01", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50+01:23", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50+0123", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50-01", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50-01:23", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},
		{str: "2013-10-08T10:30:50-0123", value: DateTimeTest{}, wantUnmarshalErr: true, noMarshal: true, tag: "no:dateTime.tz"},

		// dateTime.tz
		{str: "2013-10-08T10:30:50", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, localLoc)}, noMarshal: true},
		{str: "2013-10-08T10:30:50+01", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("+01:00", 3600))}, noMarshal: true},
		{str: "2013-10-08T10:30:50+01:23", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("+01:23", 3600+23*60))}},
		{str: "2013-10-08T10:30:50+0123", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("+01:23", 3600+23*60))}, noMarshal: true},
		{str: "2013-10-08T10:30:50-01", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("-01:00", -3600))}, noMarshal: true},
		{str: "2013-10-08T10:30:50-01:23", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("-01:23", -(3600+23*60)))}},
		{str: "2013-10-08T10:30:50-0123", value: DateTimeTzTest{time.Date(2013, 10, 8, 10, 30, 50, 0, time.FixedZone("-01:23", -(3600+23*60)))}, noMarshal: true},

		// boolean
		{str: "0", value: BooleanTest(false)},
		{str: "1", value: BooleanTest(true)},
		{str: "false", value: BooleanTest(false), noMarshal: true},
		{str: "true", value: BooleanTest(true), noMarshal: true},
		{str: "no", value: BooleanTest(false), noMarshal: true},
		{str: "yes", value: BooleanTest(true), noMarshal: true},
		{str: "", value: BooleanTest(false), noMarshal: true, wantUnmarshalErr: true},
		{str: "other", value: BooleanTest(false), noMarshal: true, wantUnmarshalErr: true},
		{str: "2", value: BooleanTest(false), noMarshal: true, wantUnmarshalErr: true},
		{str: "-1", value: BooleanTest(false), noMarshal: true, wantUnmarshalErr: true},

		// bin.base64
		{str: "", value: BinBase64Test{}},
		{str: "YQ==", value: BinBase64Test("a")},
		{str: "TG9uZ2VyIFN0cmluZy4=", value: BinBase64Test("Longer String.")},
		{str: "TG9uZ2VyIEFsaWduZWQu", value: BinBase64Test("Longer Aligned.")},

		// bin.hex
		{str: "", value: BinHexTest{}},
		{str: "61", value: BinHexTest("a")},
		{str: "4c6f6e67657220537472696e672e", value: BinHexTest("Longer String.")},
		{str: "4C6F6E67657220537472696E672E", value: BinHexTest("Longer String."), noMarshal: true},
	}

	// Generate extra test cases from convTests that implement duper.
	var extras []testCase
	for i := range tests {
		if duper, ok := tests[i].value.(duper); ok {
			duped := duper.Dupe(tests[i].tag)
			if duped != nil {
				dupedCase := testCase(tests[i])
				dupedCase.value = duped
				extras = append(extras, dupedCase)
			}
		}
	}
	tests = append(tests, extras...)

	for _, test := range tests {
		if test.noMarshal {
		} else if resultStr, err := test.value.Marshal(); err != nil && !test.wantMarshalErr {
			t.Errorf("For %T marshal %v, want %q, got error: %v", test.value, test.value, test.str, err)
		} else if err == nil && test.wantMarshalErr {
			t.Errorf("For %T marshal %v, want error, got %q", test.value, test.value, resultStr)
		} else if err == nil && resultStr != test.str {
			t.Errorf("For %T marshal %v, want %q, got %q", test.value, test.value, test.str, resultStr)
		}

		if test.noUnMarshal {
		} else if resultValue, err := test.value.Unmarshal(test.str); err != nil && !test.wantUnmarshalErr {
			t.Errorf("For %T unmarshal %q, want %v, got error: %v", test.value, test.str, test.value, err)
		} else if err == nil && test.wantUnmarshalErr {
			t.Errorf("For %T unmarshal %q, want error, got %v", test.value, test.str, resultValue)
		} else if err == nil && !test.value.Equal(resultValue) {
			t.Errorf("For %T unmarshal %q, want %v, got %v", test.value, test.str, test.value, resultValue)
		}
	}
}
