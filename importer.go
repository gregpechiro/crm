package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
)

// generic data importer
type Importer interface {
	GetHeader() []string
	GetRow(r int) []string
	GetFieldInRow(r, c int) string
	GetAllFields() [][]string
	CloseImporter()
}

// CSV File Importer...
// Implementing Importer interface
type CSVFile struct {
	fd        *os.File
	rdr       *csv.Reader
	rows      [][]string
	headerMap map[string]int
}

func NewCSVFile(path string) *CSVFile {
	c := new(CSVFile)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	c.fd = f
	c.rdr = csv.NewReader(c.fd)
	for {
		row, err := c.rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		c.rows = append(c.rows, row)
	}
	if len(c.rows) < 1 {
		log.Fatal("error reading rows")
	}
	c.headerMap = make(map[string]int)
	for i, h := range c.GetHeader() {
		c.headerMap[h] = i
	}
	return c
}

func (c *CSVFile) GetHeader() []string {
	return c.rows[0]
}

func (c *CSVFile) GetRow(r int) []string {
	if len(c.rows) < r {
		return nil
	}
	return c.rows[r]
}

func (c *CSVFile) GetFieldInRow(r, f int) string {
	if len(c.rows) < r {
		return ""
	}
	if len(c.rows[r]) < f {
		return ""
	}
	return c.rows[r][f]
}

func (c *CSVFile) GetAllFields() [][]string {
	return c.rows[1:]
}

func (c *CSVFile) CloseImporter() {
	err := c.fd.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *CSVFile) ConvertFromForm(m map[string][]string, ptr interface{}) error {
	if len(c.rows) < 2 {
		return errors.New("csv error: not enough rows in csv file")
	}
	typ := reflect.TypeOf(ptr)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("csv error: expected pointer to model\n")
	}

	// derefrencing pointer; getting model type value
	val := reflect.Indirect(reflect.ValueOf(ptr))

	// get type of single element
	typ = typ.Elem()
	strctTyp := typ.Elem()

	for rowNum := 0; rowNum < len(c.rows[1:]); rowNum++ {

		strct := reflect.Indirect(reflect.New(strctTyp))

		filled, err := c.FillStruct(rowNum, false, m, "", strct)
		if err != nil {
			return err
		}

		if filled {
			val.Set(reflect.Append(val, strct))
		}
	}
	return nil
}

func (c *CSVFile) FillStruct(rowNum int, filled bool, f map[string][]string, start string, strct reflect.Value) (bool, error) {
	for fieldNum := 0; fieldNum < strct.NumField(); fieldNum++ {
		strctTyp := strct.Type()
		fld := strct.Field(fieldNum)
		name := strctTyp.Field(fieldNum).Name
		if fld.Kind() == reflect.Struct {
			st := reflect.Indirect(fld)
			var err error
			filled, err = c.FillStruct(rowNum, filled, f, start+name+".", st)
			if err != nil {
				return false, err
			}
			fld.Set(st)
			continue
		}

		columnName, ok := getVal(start+name, f)
		if !ok {
			continue
		}

		columnNum, ok := c.headerMap[columnName]
		if !ok {
			continue
		}
		csvVal := c.GetFieldInRow(rowNum, columnNum)
		if csvVal == "" {
			continue
		}
		filled = true
		switch fld.Kind() {
		case reflect.String:
			fld.SetString(csvVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			in, err := strconv.ParseInt(csvVal, 10, 64)
			if err != nil {
				return false, errors.New(name + " Must be a a number")
			}
			fld.SetInt(in)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(csvVal, 10, 64)
			if err != nil {
				return false, errors.New(name + " Must be a a number")
			}
			fld.SetUint(u)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(csvVal, 64)
			if err != nil {
				return false, errors.New(name + " Must be a a number")
			}
			fld.SetFloat(f)
		case reflect.Bool:
			b, err := strconv.ParseBool(csvVal)
			if err != nil {
				return false, errors.New(name + " Must be either true or false")
			}
			fld.SetBool(b)
		}
	}
	return filled, nil
}

func getVal(key string, v map[string][]string) (string, bool) {
	if v == nil {
		return "", false
	}
	vs, ok := v[key]
	if !ok || len(vs) == 0 {
		return "", false
	}
	return vs[0], vs[0] != ""
}
