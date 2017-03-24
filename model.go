package main

import (
	"encoding/base64"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cagnosolutions/adb"
)

type Auth struct {
	Email    string `json:"email,omitempty" auth:"username" csv:"Email"`
	Password string `json:"password,omitempty" auth:"password" csv:"-"`
	Active   bool   `json:"active,omitempty" auth:"active" csv:"-"`
	Role     string `json:"role,omitempty" csv:"-"`
}

type Address struct {
	Street string `json:"street,omitempty"`
	City   string `json:"city,omitempty"`
	State  string `json:"state,omitempty"`
	Zip    string `json:"zip,omitempty"`
	County string `json:"county,omitempty"`
}

func (a Address) AddrHTML() string {
	if a.Street == "" && a.City == "" && a.State == "" && a.Zip == "" && a.County == "" {
		return ""
	}
	address := a.Street + "<br>" + a.City + ", "
	if a.County != "" {
		address += a.County + ", "
	}
	address += a.State + " " + a.Zip
	return address
}

type CreditCard struct {
	Number          string `json:"number,omitempty"`
	ExpirationDate  string `json:"expirationDate,omitempty"`
	ExpirationMonth int    `json:"expirationMonth,omitempty"`
	ExpirationYear  int    `json:"expirationYear,omitempty"`
	SecurityCode    string `json:"securityCode,omitempty"`
}

type Employee struct {
	Auth
	Id        string `json:"id"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Home      string `json:"home,omitempty"`
	Address
}

type Company struct {
	Auth                    `csv:"Company"`
	Id                      string     `json:"id" csv:"-"`
	DOTNum                  string     `json:"dotNum,omitempty" csv:"DOT Number"`
	Name                    string     `json:"name,omitempty" csv:"Name"`
	DBA                     string     `json:"dba,omitempty" csv:"-"`
	ContactName             string     `json:"contactName,omitempty" csv:"-"`
	ContactTitle            string     `json:"contactTitle,omitempty" csv:"-"`
	ContactSSN              string     `json:"contactSSN,omitempty" csv:"-"`
	ContactPhone            string     `json:"contactPhone,omitempty" csv:"-"`
	ContactAddress          Address    `json:"contactAddress,omitempty" csv:"-"`
	SecondName              string     `json:"secondName,omitempty" csv:"-"`
	SecondTitle             string     `json:"secondTitle,omitempty" csv:"-"`
	SecondPhone             string     `json:"secondPhone,omitempty" csv:"-"`
	SameAddress             bool       `json:"sameAddress" csv:"-"`
	PhysicalAddress         Address    `json:"pysicalAddress,omitempty" csv:"Address"`
	MailingAddress          Address    `json:"mailingAddress,omitempty" csv:"-"`
	BusinessType            string     `json:"businessType,omitempty" csv:"-"`
	BusinessTypeOther       string     `json:"businessTypeOther,omitempty" csv:"-"`
	MCNum                   string     `json:"mcNum,omitempty" csv:"-"`
	PUCNum                  string     `json:"pucNum,omitempty" csv:"-"`
	Phone                   string     `json:"phone,omitempty" csv:"Phone"`
	Fax                     string     `json:"fax,omitempty" csv:"Fax"`
	EINNum                  string     `json:"einNum,omitempty" csv:"-"`
	ARPAccountNum           string     `json:"arpAccountNum,omitempty" csv:"-"`
	CarrierType             string     `json:"carrierType,omitempty" csv:"-"`
	CarrierTypeOther        string     `json:"carrierTypeOther,omitempty" csv:"-"`
	EntityNum               string     `jaon:"entityNum,omitempty" csv:"-"`
	CreditCard              CreditCard `json:"crediCard,omitempty" csv:"-"`
	NYHutUsername           string     `json:"nyHutUsername,omitempty" csv:"-"`
	NYHutPassword           string     `json:"nyHutPassword,omitempty" csv:"-"`
	NYOscarUsername         string     `json:"nyOrcarUsername,omitempty" csv:"-"`
	NYOscarPassword         string     `json:"nyOscarUsername,omitempty" csv:"-"`
	KYUseNum                string     `json:"kyUseNum,omitempty" csv:"-"`
	NMHutUsername           string     `json:"nmHutUsername,omitempty" csv:"-"`
	NMHutPassword           string     `json:"nmHutPassword,omitempty" csv:"-"`
	DOTPin                  string     `json:"dotPin,omitempty" csv:"-"`
	MCPin                   string     `json:"mcPin,omitempty" csv:"-"`
	FMCSAUsername           string     `json:"fmcsaUsername,omitempty" csv:"-"`
	FMCSAPassword           string     `json:"fmcsaPassword,omitempty" csv:"-"`
	IRPNum                  string     `json:"irpNum,omitempty" csv:"-"`
	InsuranceCompany        string     `json:"insuranceCompany,omitempty" csv:"-"`
	InsuranceNaic           string     `json:"insuranceNaic,omitempty" csv:"-"`
	InsurancePolicyNum      string     `json:"insurancePolicyNum,omitempty" csv:"-"`
	InsuranceEffectiveDate  string     `json:"insuranceEffectiveDate,omitempty" csv:"-"`
	InsuranceExpirationDate string     `json:"insuranceExpirationDate,omitempty" csv:"-"`
	OregonNum               string     `json:"oregonNum,omiyempty" csv:"-"`
	GPSProvider             string     `json:"gpsProvider,omiyempty" csv:"-"`
	GPSUsername             string     `json:"gpsUsername,omiyempty" csv:"-"`
	GPSPassword             string     `json:"gpsPassword,omiyempty" csv:"-"`
	FuelCardProvider        string     `json:"fuelCardProvider,omiyempty" csv:"-"`
	FuelCardUsername        string     `json:"fuelCardUsername,omiyempty" csv:"-"`
	FuelCardPassword        string     `json:"fuelCardPassword,omiyempty" csv:"-"`
}

func (c Company) GetBusinessType() string {
	if c.BusinessType == "OTHER" {
		return c.BusinessTypeOther
	}
	return c.BusinessType
}

func (c Company) PasswordCheck() bool {
	return c.Password != c.Email
}

type CompanyService struct {
	Id                           string `json:"id"`
	CompanyId                    string `json:"companyId,omitempty"`
	Apportion                    bool   `json:"apportion"`
	ApportionDateOne             string `json:"apportionDateOne,omitempty"`
	ApportionOneComplete         bool   `json:"apportionOneComplete"`
	ApportionOneNotify           bool   `json:"apportionOneNotify"`
	ApportionDateTwo             string `json:"apportionDateTwo,omitempty"`
	ApportionTwoComplete         bool   `json:"apportionTwoComplete"`
	ApportionTwoNotify           bool   `json:"apportionTwoNotify"`
	FuelTaxProgram               bool   `json:"fuelTaxProgram"`
	FuelTaxProgramComplete       bool   `json:"fuelTaxProgramComplete"`
	FuelTaxProgramNotify         bool   `json:"fuelTaxProgramNotify"`
	FuelTaxNY                    bool   `json:"fuelTaxNY"`
	FuelTaxNYComplete            bool   `json:"fuelTaxNYComplete"`
	FuelTaxNYNotify              bool   `json:"fuelTaxNYNotify"`
	FuelTaxKY                    bool   `json:"fuelTaxKY"`
	FuelTaxKYComplete            bool   `json:"fuelTaxKYComplete"`
	FuelTaxKYNotify              bool   `json:"fuelTaxKYNotify"`
	FuelTaxNM                    bool   `json:"fuelTaxNM"`
	FuelTaxNMComplete            bool   `json:"FuelTaxNMComplete"`
	FuelTaxNMNotify              bool   `json:"FuelTaxNMNotify"`
	DrugConsortium               bool   `json:"drugConsortium"`
	DrugConsortiumDate           string `json:"drugConsortiumDate,omitempty"`
	DrugConsortiumComplete       bool   `json:"drugConsortiumComplete"`
	DrugConsortiumNotify         bool   `json:"drugConsortiumNotify"`
	DriverFileManagement         bool   `json:"driverFileManagement"`
	DriverFileManagementDate     string `json:"driverFileManagementDate,omitempty"`
	DriverFileManagementComplete bool   `json:"driverFileManagementComplete"`
	DriverFileManagementNotify   bool   `json:"driverFileManagementNotify"`
	DOTUpdate                    bool   `json:"dotUpdate"`
	DOTUpdateDate                string `json:"dotUpdateDate,omitempty"`
	DOTUpdateComplete            bool   `json:"dotUpdateComplete"`
	DOTUpdateNotify              bool   `json:"dotUpdateNotify"`
	TwentyTwoNinety              bool   `json:"twentyTwoNinety"`
	TwentyTwoNinetyComplete      bool   `json:"twentyTwoNinetyComplete"`
	TwentyTwoNinetyNotify        bool   `json:"twentyTwoNinetyNotify"`
	UCR                          bool   `json:"ucr"`
	UCRComplete                  bool   `json:"ucrComplete"`
	UCRNotify                    bool   `json:"ucrNotify"`
	LogAuditing                  bool   `json:"logAuditing"`
	LogAuditingComplete          bool   `json:"logAuditingComplete"`
	LogAuditingNotify            bool   `json:"logAuditingNotify"`
	CSAService                   bool   `json:"csaService"`
	CSAServiceDate               string `json:"csaServiceDate,omitempty"`
	CSAServiceComplete           bool   `json:"csaServiceComplete"`
	CSAServiceNotify             bool   `json:"csaServiceNotify"`
	NY                           bool   `json:"ny"`
	NYDate                       string `json:"nyDate"`
	NYComplete                   bool   `json:"nyComplete"`
	NYNotify                     bool   `json:"nyNotify"`
	GPS                          bool   `json:"gps"`
	GPSDate                      string `json:"gpsDate,omitempty"`
	GPSComplete                  bool   `json:"gpsComplete"`
	GPSNotify                    bool   `json:"gpsNotify"`
	Training                     bool   `json:"training"`
	TrainingDate                 string `json:"trainingDate,omitempty"`
	TrainingComplete             bool   `json:"trainingComplete"`
	TrainingNotify               bool   `json:"trainingNotify"`
	IFTARenewal                  bool   `json:"iftaRenewal"`
	IFTARenewalComplete          bool   `json:"iftaRenewalComplete"`
	IFTARenewalNotify            bool   `json:"iftaRenewalNotify"`
}

func (c *CompanyService) ResetNotifications() {
	c.ApportionOneComplete = false
	c.ApportionOneNotify = false
	c.ApportionTwoComplete = false
	c.ApportionTwoNotify = false
	c.FuelTaxProgramComplete = false
	c.FuelTaxProgramNotify = false
	c.FuelTaxNYComplete = false
	c.FuelTaxNYNotify = false
	c.FuelTaxKYComplete = false
	c.FuelTaxKYNotify = false
	c.FuelTaxNMComplete = false
	c.FuelTaxNMNotify = false
	c.DrugConsortiumComplete = false
	c.DrugConsortiumNotify = false
	c.DriverFileManagementComplete = false
	c.DriverFileManagementNotify = false
	c.DOTUpdateComplete = false
	c.DOTUpdateNotify = false
	c.TwentyTwoNinetyComplete = false
	c.TwentyTwoNinetyNotify = false
	c.UCRComplete = false
	c.UCRNotify = false
	c.LogAuditingComplete = false
	c.LogAuditingNotify = false
	c.CSAServiceComplete = false
	c.CSAServiceNotify = false
	c.NYComplete = false
	c.NYNotify = false
	c.GPSComplete = false
	c.GPSNotify = false
	c.TrainingComplete = false
	c.TrainingNotify = false
	c.IFTARenewalComplete = false
	c.IFTARenewalNotify = false
}

func (c *CompanyService) GenNotifications() {
	now := time.Now()
	beg := time.Date(now.Year(), (now.Month() + 1), now.Day(), 0, 0, 0, 0, now.Location())
	end := beg.AddDate(0, 1, 0)
	var check time.Time
	var err error
	var nId string
	var n Notification

	if c.Apportion {
		if !c.ApportionOneComplete && !c.ApportionOneNotify {
			check, err = time.Parse("01/02", c.ApportionDateOne)
			if err == nil {
				check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
				if check.After(beg) && check.Before(end) {
					// create new notification
					nId = strconv.Itoa(int(time.Now().UnixNano()))
					n = Notification{
						Id:      nId,
						ModelId: c.CompanyId,
						Type:    "COMPANY",
						SubType: "SERVICE",
						Title:   "Apportion Due",
						Body:    "Your Apportion is due on " + c.ApportionDateOne,
						Manual:  false,
					}
					db.Add("notification", nId, n)
					c.ApportionOneNotify = true
				}
			}
		}

		if !c.ApportionTwoComplete && !c.ApportionTwoNotify {
			check, err = time.Parse("01/02", c.ApportionDateTwo)
			if err == nil {
				check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
				if check.After(beg) && check.Before(end) {
					// create new notification
					nId = strconv.Itoa(int(time.Now().UnixNano()))
					n = Notification{
						Id:      nId,
						ModelId: c.CompanyId,
						Type:    "COMPANY",
						SubType: "SERVICE",
						Title:   "Apportion Due",
						Body:    "Your Apportion is due on " + c.ApportionDateTwo,
						Manual:  false,
					}
					db.Add("notification", nId, n)
					c.ApportionTwoNotify = true
				}
			}
		}

	}

	if c.FuelTaxProgram && !c.FuelTaxProgramComplete && !c.FuelTaxProgramNotify {
		check = time.Date(beg.Year(), 3, 30, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "Fuel Tax Due",
				Body:    "Your Fuel Tax is due on 03/30",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.FuelTaxProgramNotify = true
		}
	}
	if c.FuelTaxNY && !c.FuelTaxNYComplete && !c.FuelTaxNYNotify {
		check = time.Date(beg.Year(), 3, 30, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "Fuel Tax NY Due",
				Body:    "Your Fuel Tax NY is due on 03/30",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.FuelTaxNYNotify = true
		}
	}
	if c.FuelTaxNY && !c.FuelTaxKYComplete && !c.FuelTaxKYNotify {
		check = time.Date(beg.Year(), 3, 30, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "Fuel Tax KY Due",
				Body:    "Your Fuel Tax KY is due on 03/30",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.FuelTaxKYNotify = true
		}
	}
	if c.FuelTaxNM && !c.FuelTaxNMComplete && !c.FuelTaxNMNotify {
		check = time.Date(beg.Year(), 3, 30, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "Fuel Tax NM Due",
				Body:    "Your Fuel Tax NM is due on 03/30",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.FuelTaxNMNotify = true
		}
	}

	if c.DrugConsortium && !c.DrugConsortiumComplete && !c.DrugConsortiumNotify {
		check, err = time.Parse("01/02", c.DrugConsortiumDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "Drug Consortium Due",
					Body:    "Your Drug Consortium is due on " + c.DrugConsortiumDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.DrugConsortiumNotify = true
			}
		}
	}

	if c.DriverFileManagement && !c.DriverFileManagementComplete && !c.DriverFileManagementNotify {
		check, err = time.Parse("01/02", c.DriverFileManagementDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "Driver File Management Due",
					Body:    "Your Driver File Management is due on " + c.DriverFileManagementDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.DriverFileManagementNotify = true
			}
		}
	}

	if c.DOTUpdate && !c.DOTUpdateComplete && !c.DOTUpdateNotify {
		check, err = time.Parse("01/02", c.DOTUpdateDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "DOT Update Due",
					Body:    "Your DOT Update is due on " + c.DOTUpdateDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.DOTUpdateNotify = true
			}
		}
	}

	if c.TwentyTwoNinety && !c.TwentyTwoNinetyComplete && !c.TwentyTwoNinetyNotify {
		check = time.Date(beg.Year(), 6, 1, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "2290 Due",
				Body:    "Your 2290 is due on 06/01",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.TwentyTwoNinetyNotify = true
		}
	}

	if c.UCR && !c.UCRComplete && !c.UCRNotify {
		check = time.Date(beg.Year(), 10, 1, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "UCR Due",
				Body:    "Your UCR is due on 10/01",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.UCRNotify = true
		}
	}

	if c.LogAuditing && !c.LogAuditingComplete && !c.LogAuditingNotify {
		check = time.Date(beg.Year(), 1, 7, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "Log Auditing Due",
				Body:    "Your Log Auditing is due on 01/07",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.LogAuditingNotify = true
		}
	}

	if c.CSAService && !c.CSAServiceComplete && !c.CSAServiceNotify {
		check, err = time.Parse("01/02", c.CSAServiceDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "CSA Service Due",
					Body:    "Your CSA Service is due on " + c.CSAServiceDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.CSAServiceNotify = true
			}
		}
	}

	if c.IFTARenewal && !c.IFTARenewalComplete && !c.IFTARenewalNotify {
		check = time.Date(beg.Year(), 11, 15, 0, 0, 0, 0, beg.Location())
		check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			// create new notification
			nId = strconv.Itoa(int(time.Now().UnixNano()))
			n = Notification{
				Id:      nId,
				ModelId: c.CompanyId,
				Type:    "COMPANY",
				SubType: "SERVICE",
				Title:   "IFTA Renewal Due",
				Body:    "Your IFTA Rewnal is due on 11/15",
				Manual:  false,
			}
			db.Add("notification", nId, n)
			c.IFTARenewalNotify = true
		}
	}

	if c.NY && !c.NYComplete && !c.NYNotify {
		check, err = time.Parse("01/02", c.NYDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "NY Due",
					Body:    "Your NY is due on " + c.NYDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.NYNotify = true
			}
		}
	}

	if c.GPS && !c.GPSComplete && !c.GPSNotify {
		check, err = time.Parse("01/02", c.GPSDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "GPS Due",
					Body:    "Your GPS is due on " + c.GPSDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.GPSNotify = true
			}
		}
	}

	if c.Training && !c.TrainingComplete && !c.TrainingNotify {
		check, err = time.Parse("01/02", c.TrainingDate)
		if err == nil {
			check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
			if check.After(beg) && check.Before(end) {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: c.CompanyId,
					Type:    "COMPANY",
					SubType: "SERVICE",
					Title:   "Training Due",
					Body:    "Your Training is due on " + c.TrainingDate,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				c.TrainingNotify = true
			}
		}
	}

}

type CompanyServiceEmails struct {
	Id                         string `json:"id"`
	CompanyId                  string `json:"companyId"`
	ApportionOneEmailId        string `json:"apportionOneEmailId,omitempty"`
	ApportionTwoEmailId        string `json:"apportionTwoEmailId,omitempty"`
	FuelTaxProgramEmailId      string `json:"fuelTaxProgramEmailId,omitempty"`
	FuelTaxNYEmailId           string `json:"fuelTaxNYEmailId,omitempty"`
	FuelTaxKYEmailId           string `json:"fuelTaxKYEmailId,omitempty"`
	FuelTaxNMEmailId           string `json:"fuelTaxNMEmailId,omitempty"`
	DrugConsortiumEmailId      string `json:"drugConsortiumEmailId,omitempty"`
	DriverFileManagmentEmailId string `json:"driverFileManagmentEmailId,omitempty"`
	DOTUpdateEmailId           string `json:"dotUpdateEmailId,omitempty"`
	TwentyTwoNinetyEmailId     string `json:"twentyTwoNinetyEmailId,omitempty"`
	UCREmailId                 string `json:"ucrEmailId,omitempty"`
	LogAuditingEmailId         string `json:"logAuditingEmailId,omitempty"`
	CSAServiceEmailId          string `json:"csaServiceEmailId,omitempty"`
	NYEmailId                  string `json:"nyEmailId,omitempty"`
	GPSEmailId                 string `json:"gpsEmailId,omitempty"`
	TrainingEmailId            string `json:"trainingEmailId,omitempty"`
	IFTARenewalEmailId         string `json:"iftaRenewalemailId,omitempty"`
}

type CompanyFeatures struct {
	Id        string `json:"id"`
	CompanyId string `json:"companyId"`
	Vehicles  bool   `json:"vehicles"`
	Drivers   bool   `json:"drivers"`
	Forms     bool   `json:"forms"`
	Files     bool   `json:"files"`
}

type Driver struct {
	Auth
	Address
	Id                     string `json:"id"`
	EmployeeId             string `json:"employeeId,omitempty"`
	OriginalId             string `json:"originalId,omitempty"`
	FirstName              string `json:"firstName,omitempty"`
	LastName               string `json:"lastName,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	EmergencyContactName   string `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone  string `json:"emergencyContactPhone,omitempty"`
	LicenseNum             string `json:"licenseNum,omitempty"`
	DOB                    string `json:"dob,omitempty"`
	LicenseState           string `json:"licenseState,omitempty"`
	LicenseExpire          string `json:"licenseExpire,omitempty"`
	LicenseExpireEmailId   string `json:"licenseExpireEmailId, omitempty"`
	LicenseExpireNotify    bool   `json:"licenseExpireNotify"`
	MedCardExpiry          string `json:"medCardExpiry,omitempty"`
	MedCardExpireEmailId   string `json:"medCardExpireEmailId, omitempty"`
	MedCardExpireNotify    bool   `json:"medCardExpireNotify"`
	MVRExpiry              string `json:"mVRExpiry,omitempty"`
	MVRExpireEmailId       string `json:"mVRExpireEmailId, omitempty"`
	MVRExpireNotify        bool   `json:"mVRExpireNotify"`
	ReviewExpiry           string `json:"reviewExpiry,omitempty"`
	ReviewExpireEmailId    string `json:"reviewExpireEmailId, omitempty"`
	ReviewExpireNotify     bool   `json:"reviewExpireNotify"`
	OneEightyExpiry        string `json:"oneEightyExpiry,omitempty"`
	OneEightyExpireEmailId string `json:"oneEightyExpireEmailId, omitempty"`
	OneEightyExpireNotify  bool   `json:"oneEightyExpireNotify"`
	HireDate               string `json:"hireDate,omitempty"`
	TermDate               string `json:"termDate,omitempty"`
	CompanyId              string `json:"companyId,omitempty"`
	Status                 string `json:"driverStatus,omitempty"`
}

func DeleteDriver(driverId string) {
	var documents []Document
	db.TestQuery("document", &documents, adb.Eq("driverId", `"`+driverId+`"`))
	for _, doc := range documents {
		db.Del("document", doc.Id)
	}

	os.RemoveAll("upload/driver/" + driverId + "/")

	db.Del("driver", driverId)
}

func (d *Driver) GenNotifications() {
	now := time.Now()
	beg := time.Date(now.Year(), (now.Month() + 1), now.Day(), 0, 0, 0, 0, now.Location())
	end := beg.AddDate(0, 1, 0)
	var check time.Time
	var err error
	var nId string
	var n Notification

	// license expire
	check, err = time.Parse("1/2/2006", d.LicenseExpire)
	if err == nil {
		//check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			if !d.LicenseExpireNotify {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: d.CompanyId,
					Type:    "COMPANY",
					SubType: "DRIVER-FORM",
					Title:   "License Expiring",
					Body:    d.FirstName + " " + d.LastName + "'s license is expiring on " + d.LicenseExpire,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				d.LicenseExpireNotify = true
			}
		} else {
			d.LicenseExpireNotify = false
		}
	}

	// med card expire
	check, err = time.Parse("1/2/2006", d.MedCardExpiry)
	if err == nil {
		//check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			if !d.MedCardExpireNotify {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: d.CompanyId,
					Type:    "COMPANY",
					SubType: "DRIVER-FORM",
					Title:   "Medical Card Expiring",
					Body:    d.FirstName + " " + d.LastName + "'s medical card is expiring on " + d.MedCardExpiry,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				d.MedCardExpireNotify = true
			}
		} else {
			d.MedCardExpireNotify = false
		}
	}

	// mvr expire
	check, err = time.Parse("1/2/2006", d.MVRExpiry)
	if err == nil {
		//check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			if !d.MVRExpireNotify {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: d.CompanyId,
					Type:    "COMPANY",
					SubType: "DRIVER-FORM",
					Title:   "MVR Expiring",
					Body:    d.FirstName + " " + d.LastName + "'s MVR is expiring on " + d.MVRExpiry,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				d.MVRExpireNotify = true
			}
		} else {
			d.MVRExpireNotify = false
		}
	}

	// review expire
	check, err = time.Parse("1/2/2006", d.ReviewExpiry)

	if err == nil {
		//check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			if !d.ReviewExpireNotify {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: d.CompanyId,
					Type:    "COMPANY",
					SubType: "DRIVER-FORM",
					Title:   "Review due",
					Body:    d.FirstName + " " + d.LastName + "'s review is due on " + d.ReviewExpiry,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				d.ReviewExpireNotify = true
			}
		} else {
			d.ReviewExpireNotify = false
		}
	}

	// 180 expire
	check, err = time.Parse("1/2/2006", d.OneEightyExpiry)

	if err == nil {
		//check = check.AddDate(beg.Year(), 0, 0).In(beg.Location())
		if check.After(beg) && check.Before(end) {
			if !d.OneEightyExpireNotify {
				// create new notification
				nId = strconv.Itoa(int(time.Now().UnixNano()))
				n = Notification{
					Id:      nId,
					ModelId: d.CompanyId,
					Type:    "COMPANY",
					SubType: "DRIVER-FORM",
					Title:   "180 Expiring",
					Body:    d.FirstName + " " + d.LastName + "'s 180 is expiring on " + d.OneEightyExpiry,
					Manual:  false,
				}
				db.Add("notification", nId, n)
				d.OneEightyExpireNotify = true
			}
		} else {
			d.OneEightyExpireNotify = false
		}
	}

}

func (d Driver) GetAge() int {
	dobT, err := time.Parse("1/2/2006", d.DOB)
	if err != nil {
		return 0
	}
	dob := dobT.UnixNano()
	diff := time.Now().UnixNano() - dob
	return int(math.Floor((float64(diff) / float64(1000) / float64(1000) / float64(1000) / float64(60) / float64(60) / float64(24) / float64(365.25))))
}

type Vehicle struct {
	Id               string  `json:"id"`
	CompanyId        string  `json:"companyId,omitempty"`
	VehicleType      string  `json:"vehicleType,omitempty"`
	UnitNumber       string  `json:"unitNumber,omitempty"`
	Make             string  `json:"make,omitempty"`
	VIN              string  `json:"vin,omitempty"`
	Title            string  `json:"title,omitempty"`
	GVW              int     `json:"gvw,omitempty"`
	GCR              int     `json:"gcr,omitempty"`
	UnladenWeight    int     `json:"unladenWeight,omitempty"`
	PurchasePrice    float32 `json:"purchasePrice,omitempty"`
	PurchaseDate     string  `json:"purchaseDate,omitempty"`
	CurrentValue     float32 `json:"currentValue,omitempty"`
	AxleAmount       string  `json:"axleAmount,omitempty"`
	FuelType         string  `json:"fuelType,omitempty"`
	FuelTypeOther    string  `json:"fuelTypeOther,omitempty"`
	Active           bool    `json:"active"`
	Owner            string  `json:"owner,omitempty"`
	Year             string  `json:"year,omitempty"`
	PlateNum         string  `json:"plateNum,omitempty"`
	PlateExpire      string  `json:"plateExpire,omitempty"`
	PlateExpireMonth string  `json:"plateExpireMonth,omitempty"`
	PlateExpireYear  string  `json:"plateExpireYear,omitempty"`
	BodyType         string  `json:"bodyType,omitempty"`
	BodyTypeOther    string  `json:"bodyTypeOther,omitempty"`
}

func DeleteVehicle(vehicleId string) {
	os.RemoveAll("upload/vehicle/" + vehicleId + "/")
	db.Del("vehicle", vehicleId)
}

func (v Vehicle) HigherWeight() int {
	if v.GVW > v.GCR {
		return v.GVW
	}
	return v.GCR
}

func (v Vehicle) GetBodyType() string {
	if v.BodyType == "O" {
		return v.BodyTypeOther
	}
	return v.BodyType
}

func (v Vehicle) GetFuelType() string {
	switch v.FuelType {
	case "D":
		return "Diesel"
	case "G":
		return "Gas"
	case "H":
		return "Hybrid"
	case "N":
		return "Natural Gas"
	case "P":
		return "Propane"
	case "O":
		return v.FuelTypeOther
	}
	return "NONE"
}

type Note struct {
	Id              string `json:"id,omitempty"`
	CompanyId       string `json:"companyId,omitempty"`
	EmployeeId      string `json:"employeeId,omitempty"`
	Communication   string `json:"communication,omitempty"`
	Purpose         string `json:"purpose,omitempty"`
	StartTime       int64  `json:"startTime,omitempty"`
	StartTimePretty string `json:"startTimePretty,omitempty"`
	EndTime         int64  `json:"endTime,omitempty"`
	EndTimePretty   string `json:"endTimePretty,omitempty"`
	Representative  string `json:"representative,omitempty"`
	CallBack        string `json:"callBack,omitempty"`
	EmailEmployee   bool   `json:"emailEmployee,omitempty"`
	Billable        bool   `json:"billable,omitempty"`
	Body            string `json:"body,omitempty"`
}

type NoteSort []Note

func (n NoteSort) Len() int {
	return len(n)
}

func (n NoteSort) Less(i, j int) bool {
	return n[i].StartTime < n[j].StartTime
}

func (n NoteSort) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type Comment struct {
	Id     string `json:"id"`
	Body   string `json:"body"`
	Url    string `json:"url"`
	Page   string `json:"page"`
	Closed bool   `json:"closed"`
}

type QuickNote struct {
	Name string
	Body string
}

type Document struct {
	Id         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	DocumentId string   `json:"documentId,omitempty"`
	Complete   bool     `json:"complete"`
	Data       string   `json:"data,omitempty"`
	CompanyId  string   `json:"companyId,omitempty"`
	DriverId   string   `json:"driverId,omitempty"`
	VehicleIds []string `json:"vehicleIds,omitempty"`
	StateForm  bool     `json:"stateForm,omitempty"`
}

type Notification struct {
	Id      string `json:"id"`
	ModelId string `json:"modelId,omitempty"`
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	Manual  bool   `json:"manual"`
}

var DQFS = [][]string{
	[]string{"100", "Driver's Application"},
	[]string{"180", "Certification of Violations"},
	[]string{"200", "Annual Inquery & Review"},
	[]string{"250", "Road Test Certication"},
	[]string{"300", "Previous Driver Inquires"},
	[]string{"400", "Drug & Alcohol Records Request"},
	[]string{"425", "Drug & Alcohol Pre-Employment Statement"},
	[]string{"475", "Alcohol and/or Drug Test Notification"},
	[]string{"450", "Drug & Alcohol Certified Receipt"},
	[]string{"500", "Certification Compliance"},
	[]string{"600", "Confictions for a Driver Violation"},
	[]string{"700", "New Hire Stmt On Duty Hours"},
	[]string{"750", "Other Ompensated Work"},
	[]string{"775", "Fair Credit Reporting Act"},
}

var CompanyForms = [][]string{
	[]string{"MV-550", "2", ""},
	[]string{"MV-550A", "1", ""},
	//[]string{"MV-551", "2"},
	[]string{"MV-552A", "ALL", ""},
	[]string{"MV-558", "4", "none"},
	[]string{"MV-41", "1", ""},
	[]string{"TMT-39", "", ""},
	[]string{"PUC App", "", ""},
	//[]string{"MCS-150", "", ""},
}

type EmailTemplate struct {
	Id   string
	Name string
	Body string
}

type ViolationCache struct {
	LastUpdate           string `json:"lastUpdate,omitempty"`
	Cache                bool   `json:"cache,omitempty"`
	UnsafeDriving        string `json:"unsafeDriving,omitempty"`
	HOSCompliance        string `json:"hosCompliance,omitempty"`
	VehicleMaintenance   string `json:"vehicleMaintenance,omitempty"`
	ControlledSubstances string `json:"controlledSubstances,omitempty"`
	DriverFitness        string `json:"driverFitness,omitempty"`
}

func GetCustomerViolations(id string) ViolationCache {
	var company Company
	db.Get("company", id, &company)
	var violations ViolationCache
	var resp *http.Response
	var err error
	var b []byte
	hasError := false

	if !violations.NeedsUpdate() {
		return violations
	}

	resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/UnsafeDriving/Violations/--SORTBY--.aspx")
	if err != nil {
		hasError = true
		violations.UnsafeDriving = "<tr><td>Error retrieving violations</td></tr>"

	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hasError = true
		violations.UnsafeDriving = "<tr><td>Error retrieving violations</td></tr>"

	}

	violations.UnsafeDriving = strings.Replace(strings.Replace(strings.Replace(string(b), "  ", "", -1), "\n", "", -1), "\t", "", -1)
	if !strings.Contains(violations.UnsafeDriving, "violSummary") {
		violations.UnsafeDriving = "<tr><td>No violations to display.</td></tr>"
	}

	resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/DrugsAlcohol/Violations/--SORTBY--.aspx")
	if err != nil {
		hasError = true
		violations.ControlledSubstances = "<tr><td>Error retrieving violations</td></tr>"

	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hasError = true
		violations.ControlledSubstances = "<tr><td>Error retrieving violations</td></tr>"

	}

	violations.ControlledSubstances = strings.Replace(strings.Replace(strings.Replace(string(b), "  ", "", -1), "\n", "", -1), "\t", "", -1)
	if !strings.Contains(violations.ControlledSubstances, "violSummary") {
		violations.ControlledSubstances = "<tr><td>No violations to display.</td></tr>"
	}

	resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/DriverFitness/Violations/--SORTBY--.aspx")
	if err != nil {
		hasError = true
		violations.DriverFitness = "<tr><td>Error retrieving violations</td></tr>"

	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hasError = true
		violations.DriverFitness = "<tr><td>Error retrieving violations</td></tr>"

	}

	violations.DriverFitness = strings.Replace(strings.Replace(strings.Replace(string(b), "  ", "", -1), "\n", "", -1), "\t", "", -1)
	if !strings.Contains(violations.DriverFitness, "violSummary") {
		violations.DriverFitness = "<tr><td>No violations to display.</td></tr>"
	}

	resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/HOSCompliance/Violations/--SORTBY--.aspx")
	if err != nil {
		hasError = true
		violations.HOSCompliance = "<tr><td>Error retrieving violations</td></tr>"

	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hasError = true
		violations.HOSCompliance = "<tr><td>Error retrieving violations</td></tr>"

	}

	violations.HOSCompliance = strings.Replace(strings.Replace(strings.Replace(string(b), "  ", "", -1), "\n", "", -1), "\t", "", -1)
	if !strings.Contains(violations.HOSCompliance, "violSummary") {
		violations.HOSCompliance = "<tr><td>No violations to display.</td></tr>"
	}

	resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/VehicleMaint/Violations/--SORTBY--.aspx")
	if err != nil {
		hasError = true
		violations.VehicleMaintenance = "<tr><td>Error retrieving violations</td></tr>"

	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hasError = true
		violations.VehicleMaintenance = "<tr><td>Error retrieving violations</td></tr>"

	}

	violations.VehicleMaintenance = strings.Replace(strings.Replace(strings.Replace(string(b), "  ", "", -1), "\n", "", -1), "\t", "", -1)
	if !strings.Contains(violations.VehicleMaintenance, "violSummary") {
		violations.VehicleMaintenance = "<tr><td>No violations to display.</td></tr>"
	}

	violations.LastUpdate = time.Now().Format("01/02/2006")

	if !hasError {
		// NOTE: save cache here
	}

	return violations
}

func (v ViolationCache) NeedsUpdate() bool {
	t, err := time.Parse("01/02/2006", v.LastUpdate)
	if err != nil {
		return true
	}
	now := time.Now()
	beg := now.AddDate(0, 0, -7)
	return !(t.After(beg) && t.Before(now))
}

type SaferCache struct {
	LastUpdate      string `json:"lastUpdate,omitempty"`
	Cache           bool   `json:"cache,omitempty"`
	InspectionsHead string `json:"inspectionsHead,omitempty"`
	InspectionsBody string `json:"inspectionsBody,omitempty"`
	CrashesHead     string `json:"crashesHead,omitempty"`
	CrashesBody     string `json:"crashesBody,omitempty"`
}

func GetCustomerSafer(id string) SaferCache {
	var company Company
	db.Get("company", id, &company)
	var safer SaferCache

	exists := db.Get("safer-cache", id, &safer)
	if exists && !safer.NeedsUpdate() {
		safer.Cache = true
		return safer
	}

	resp, err := http.Get("https://safer.fmcsa.dot.gov/query.asp?searchtype=ANY&query_type=queryCarrierSnapshot&query_param=USDOT&query_string=" + company.DOTNum)
	if err != nil {
		safer.Cache = exists
		return safer
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		safer.Cache = exists
		return safer
	}

	safer.Cache = false
	safer.LastUpdate = time.Now().Format("01/02/2006")
	safer.InspectionsBody = base64.StdEncoding.EncodeToString(b)
	return safer
}

func (s SaferCache) NeedsUpdate() bool {
	t, err := time.Parse("01/02/2006", s.LastUpdate)
	if err != nil {
		return true
	}
	now := time.Now()
	beg := now.AddDate(0, 0, -7)
	return !(t.After(beg) && t.Before(now))
}

type Task struct {
	Id            string `json:"id"`
	EmployeeId    string `json:"employeeId,omitempty"`
	CompanyId     string `json:"companyId,omitempty"`
	CreatedTime   int64  `json:"createdTime,omitempty"`   // time.Time.Unix()
	AssignedTime  int64  `json:"assignedTime,omitempty"`  // time.Time.Unix()
	StartedTime   int64  `json:"startedTime,omitempty"`   // time.Time.Unix()
	CompletedTime int64  `json:"completedTime,omitempty"` // time.Time.Unix()
	Complete      bool   `json:"complete"`
	Description   string `json:"description,omitempty"`
	Notes         string `json:"notes,omitempty"`
	EmployeeName  string `json:"employeeName, omitempty"`
	CompanyName   string `json:"companyName, omitempty"`
}

func GetTaskEmployeeView(tasks []Task) {
	for i, task := range tasks {
		var company Company
		db.Get("company", task.CompanyId, &company)
		task.CompanyName = company.Name
		tasks[i] = task
	}
}

func GetTaskCompanyView(tasks []Task) {
	for i, task := range tasks {
		var employee Employee
		db.Get("employee", task.EmployeeId, &employee)
		task.EmployeeName = employee.FirstName + " " + employee.LastName
		tasks[i] = task
	}
}

func GetTaskAdminView(tasks []Task) {
	for i, task := range tasks {
		var company Company
		db.Get("company", task.CompanyId, &company)
		task.CompanyName = company.Name
		var employee Employee
		db.Get("employee", task.EmployeeId, &employee)
		task.EmployeeName = employee.FirstName + " " + employee.LastName
		tasks[i] = task
	}
}

type Node struct {
	Id       string `json:"id"`
	Text     string `json:"text"`
	Children bool   `json:"children"`
	Type     string `json:"type"`
	State    string `json:"state"`
}
