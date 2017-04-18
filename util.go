package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func ImportSave(v interface{}) {
	switch v.(type) {
	case []Employee:
		employees := v.([]Employee)
		for _, employee := range employees {
			employee.Id = strconv.Itoa(int(time.Now().UnixNano()))
			db.Set("employee", employee.Id, employee)
		}

	case []Customer:
		customer := v.([]Customer)
		for _, customer := range customer {
			customer.Id = strconv.Itoa(int(time.Now().UnixNano()))
			db.Set("customer", customer.Id, customer)
		}
	}
}

var CSVSINGLE = map[string]interface{}{
	"customer": Customer{},
	"employee": Employee{},
}

// setup quick notes
var quickNotes = []QuickNote{
	// QuickNote{"Complete", "Completed task."},
	// QuickNote{"Updated", "Information updated."},
	QuickNote{"Okay", "I had okay contact with the customer."},
	QuickNote{"Happy", "The customer was happy when I was finished with them."},
	QuickNote{"Sad", "The customer was sad when I was finished with them."},
	QuickNote{"Mad", "The customer was mad when I was finished with them."},
}

func IsEmptyDir(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err != nil {
		return true
	}
	return false
}

func Today() (int64, int64) {
	loc, _ := time.LoadLocation("Local")
	now := time.Now()
	beg := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := beg.AddDate(0, 0, 1)
	return beg.Unix() - 1, end.Unix()
}

func ThisMonth() (int64, int64) {
	loc, _ := time.LoadLocation("Local")
	now := time.Now()
	beg := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	end := beg.AddDate(0, 1, 0)
	return beg.Unix() - 1, end.Unix()
}

func PrettySize(size int64) string {
	c := 0
	var sizef float64 = float64(size)
	for sizef > 1024 {
		sizef = sizef / 1024
		c++
	}
	ind := ""
	switch c {
	case 0:
		ind = "B"
	case 1:
		ind = "KB"
	case 2:
		ind = "MB"
	case 3:
		ind = "GB"
	}
	return fmt.Sprintf("%.1f %s", sizef, ind)
}

func PrintTime(tt int64) string {
	h := tt / 3600
	m := (tt % 3600) / 60

	if h > 0 || m > 0 {
		return fmt.Sprintf("%d hour(s), %d minute(s)", h, m)
	}
	return ""
}

func ajaxResponse(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, msg)
}

func FormatDate(d string) string {
	ds := strings.Split(d, "-")
	if len(ds) != 3 {
		return ""
	}
	if ds[1][0] == '0' {
		ds[1] = ds[1][1:]
	}
	return fmt.Sprintf("%s/%s/%s", ds[1], ds[2], ds[0])
}

func ToJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func ToBase64Json(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func IdTime(id string) string {
	idT, err := strconv.Atoi(id)
	if err != nil || idT == 0 {
		return ""
	}
	t := time.Unix(0, int64(idT))
	return t.Format("01/02/2006 03:04 PM")
}

func PrettyDate(ts int64) string {
	if ts == 0 {
		return ""
	}
	t := time.Unix(ts, 0)
	return t.Format("1/2/2006")
}

func PrettyDateTime(ts int64) string {
	if ts == 0 {
		return ""
	}
	t := time.Unix(ts, 0)
	return t.Format("1/02/2006 03:04 PM")
}

func add(i, j int) int {
	return i + j
}

func CopyDir(source string, dest string) error {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	if err := os.MkdirAll(dest, sourceinfo.Mode()); err != nil {
		return err
	}

	dir, _ := os.Open(source)

	objects, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {

		sourceFile := source + "/" + obj.Name()

		destFile := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourceFile, destFile)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = CopyFile(sourceFile, destFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(source string, dest string) error {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err != nil {
		return err
	}
	sourceinfo, err := os.Stat(source)
	if err != nil {
		if err := os.Chmod(dest, sourceinfo.Mode()); err != nil {
			return err
		}
	}

	return nil
}

/*func FormToStruct(ptr interface{}, vals map[string][]string, form string) (map[string]string, bool) {
	errors := make(map[string]string)
	formToStruct(ptr, vals, "", errors, form)
	return errors, len(errors) == 0
}

func formToStruct(ptr interface{}, vals map[string][]string, start string, errors map[string]string, form string) {
	var strct reflect.Value
	if reflect.TypeOf(ptr) == reflect.TypeOf(reflect.Value{}) {
		strct = ptr.(reflect.Value)
	} else {
		strct = reflect.ValueOf(ptr).Elem()
	}
	strctType := strct.Type()
	for i := 0; i < strct.NumField(); i++ {
		fld := strct.Field(i)
		name := ToLowerFirst(strctType.Field(i).Name)
		if ok, v := GetVal(start+name, vals); ok || fld.Kind() == reflect.Struct {
			if fld.Kind() != reflect.Struct && v == "" && strings.Index(string(strctType.Field(i).Tag), "required") == 0 {
				errors[start+name] = ToUpperFirst(name) + " is required"
			}
			switch fld.Kind() {
			case reflect.String:
				fld.SetString(v)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				in, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					errors[start+name] = ToUpperFirst(name) + " must be a number"
				}
				fld.SetInt(in)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				u, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					errors[start+name] = ToUpperFirst(name) + " must be a number"
				}
				fld.SetUint(u)
			case reflect.Float32, reflect.Float64:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					errors[start+name] = ToUpperFirst(name) + " must be a number"
				}
				fld.SetFloat(f)
			case reflect.Bool:
				b, err := strconv.ParseBool(v)
				if err != nil {
					errors[start+name] = ToUpperFirst(name) + " must be either true or false"
				}
				fld.SetBool(b)
			case reflect.Slice:
				ss := reflect.MakeSlice(fld.Type(), 0, 0)
				fld.Set(genSlice(ss, v, start, name, errors))
			case reflect.Struct:
				st := reflect.Indirect(fld)
				formToStruct(st, vals, start+name+".", errors, form)
				fld.Set(st)
			}
		}
	}
}

func genSlice(sl reflect.Value, val, start, name string, errors map[string]string) reflect.Value {
	vs := strings.Split(val, ",")
	for _, v := range vs {
		switch sl.Type().String() {
		case "[]string":
			sl = reflect.Append(sl, reflect.ValueOf(v))
		case "[]int":
			in, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(int(in)))
		case "[]int8":
			in, err := strconv.ParseInt(v, 10, 8)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(int8(in)))
		case "[]int16":
			in, err := strconv.ParseInt(v, 10, 16)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(int16(in)))
		case "[]int32":
			in, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(int32(in)))
		case "[]int64":
			in, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(int64(in)))
		case "[]uint":
			in, err := strconv.ParseUint(v, 10, 0)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(uint(in)))
		case "[]uint8":
			in, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(uint8(in)))
		case "[]uint16":
			in, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(uint16(in)))
		case "[]uint32":
			in, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(uint32(in)))
		case "[]uint64":
			in, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(uint64(in)))
		case "[]float32":
			in, err := strconv.ParseFloat(v, 32)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(float32(in)))
		case "[]float64":
			in, err := strconv.ParseFloat(v, 64)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of numbers"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(float64(in)))
		case "[]bool":
			b, err := strconv.ParseBool(v)
			if err != nil {
				errors[start+name] = ToUpperFirst(name) + " must be a list of either true or false"
				break
			}
			sl = reflect.Append(sl, reflect.ValueOf(b))
		}
	}
	return sl
}

func GetVal(key string, v map[string][]string) (bool, string) {
	if v == nil {
		return false, ""
	}
	vs, ok := v[key]
	if !ok || len(vs) == 0 {
		return false, ""
	}
	return true, vs[0]
}

func ToLowerFirst(s string) string {
	return strings.ToLower(string(s[0])) + s[1:len(s)]
}

func ToUpperFirst(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:len(s)]
}*/
