package main

import (
        "crypto/tls"
        "encoding/json"
        "fmt"
        "io/ioutil"
        "net/http"
        "os"
        "strconv"
        "strings"
        "time"

        "gopkg.in/gomail.v1"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////// CONSTANTES ////////////////////////////////////////////////////////

var arrayJSON Response
var object objectType
var expirations []objectType

var objectName string
var daysRemaining uint32
var contractEnd string
var snValue string

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////// OBJETS ///////////////////////////////////////////////////

type Response struct {
        Response map[string]interface{} `json:"response"`
}

type objectType struct {
        name string
        sn   string
        end  string
        days uint32
}

type expirationsBox struct {
        object []objectType
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////// FONCTIONS ////////////////////////////////////////////////

func httpGetWithAuth() {
        tr := &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        client := &http.Client{Transport: tr}
        req, err := http.NewRequest("GET", "https://SERVERNAME/api.php?method=get_depot&include_attrs", nil)
        req.SetBasicAuth("user", "mdp")
        resp, err := client.Do(req)
        if err != nil {
                fmt.Printf("Error : %s", err)
        }
        defer resp.Body.Close()
        body, _ := ioutil.ReadAll(resp.Body)
        data := string(body)
        err = json.Unmarshal([]byte(data), &arrayJSON)
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
        return
}

func mapRequestParsing() {
        for _, objects := range arrayJSON.Response {
                object := objects.(map[string]interface{})
                objectName = object["name"].(string)
                if strings.Contains(objectName, "prd") {
                        dateValue := getValueFromAttr(object, "Support contract expiration")
                        snValue = getValueFromAttr(object, "OEM S/N 1")
                        dataWork(dateValue)
                } else {
                        continue
                }
                jeVeuxDeLOrdre()
        }
        return
}

func getValueFromAttr(object map[string]interface{}, attrName string) (value string) {
        attrs := object["attrs"].(map[string]interface{})
        value = "Non renseigné"
        for _, attr := range attrs {
                eachAttr := attr.(map[string]interface{})
                if eachAttr["name"] == attrName {
                        value = eachAttr["value"].(string)
                }
        }
        return
}

func jeVeuxDeLOrdre() {
        var arrayZero []objectType
        var arrayElse []objectType
        for _, record := range expirations {
                if record.days == 0 {
                        arrayZero = append(arrayZero, record)
                } else {
                        arrayElse = append(arrayElse, record)
                }
        }
        for _, object := range arrayElse {
                arrayZero = append(arrayZero, object)
        }
        expirations = arrayZero
}

func dataWork(dateValue string) {
        if dateValue == "Non renseigné" {
                contractEnd = "Non renseigné"
                daysRemaining = 0
        } else {
                contractEnd, daysRemaining = remainingDaysNContract(dateValue)
                sortDate()
        }
        return
}

func remainingDaysNContract(valueDate string) (contractEnd string, daysRemaining uint32) {
        date64, err := strconv.ParseInt(valueDate, 10, 64)
        if err != nil {
                panic(err)
        }
        contractEnd2 := time.Unix(date64, 0)
        diffTime := time.Now().Sub(contractEnd2)
        if diffTime > 0 {
                daysRemaining = 0
                contractEnd = "Expiré"
        } else {
                daysRemaining = uint32(diffTime.Hours() * (-0.04166666))
                contractEnd = time.Unix(date64, 0).String()
                contractEnd = contractEnd[0 : len(contractEnd)-19]
        }
        return
}

func sortDate() {
        if daysRemaining <= 0 || daysRemaining == 30 || daysRemaining == 60 || daysRemaining == 90 {
                object.name = objectName
                object.sn = snValue
                object.end = contractEnd
                object.days = daysRemaining
                expirations = append(expirations, object)
        }
        return
}

func tableauHTML() string {
        var mailBody string
        mailBody = fmt.Sprintf("<html><head><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"></head><body><br><br><center>")
        mailBody = fmt.Sprintf("%s<h1 style=\"margin-bottom:-20px; color:#4F81BD; border:2px;\"><u>Support contract expirations</u></h1>\n", mailBody)
        mailBody = fmt.Sprintf("%s<br><br><br><table border=\"3\" bordercolor=#4F81BD cellpadding=\"10\" cellspacing=\"0\"><tr style=\"background-color:#4F81BD; color:white;\">", mailBody)
        mailBody = fmt.Sprintf("%s<th>Nom</th><th>Numéro de série</th><th>Fin de contrat</th><th>Jours restants</th></tr>\n", mailBody)
        mailBody = objectRow(mailBody)
        mailBody = fmt.Sprintf("%s</table></center></body></html>", mailBody)
        return (mailBody)
}

func objectRow(mailBody string) string {
        for _, record := range expirations {
                mailBody = fmt.Sprintf("%s<tr style=\"background-color: #D0D8E8\"><td align=\"center\" style=\"text-decoration:none;color:#4F81BD;\">%s</td>", mailBody, record.name)
                mailBody = fmt.Sprintf("%s<td align=\"center\">%s</td>", mailBody, record.sn)
                mailBody = fmt.Sprintf("%s<td align=\"center\">%s</td>", mailBody, record.end)
                if record.days == 0 {
                        mailBody = fmt.Sprintf("%s<td align=\"center\"style=\"background-color: #FF8E95\">%d</td></tr>", mailBody, record.days)
                } else if record.days == 30 {
                        mailBody = fmt.Sprintf("%s<td align=\"center\"style=\"background-color: #FF7F20\">%d</td></tr>", mailBody, record.days)
                } else if record.days == 60 {
                        mailBody = fmt.Sprintf("%s<td align=\"center\"style=\"background-color: #FDFEB2\">%d</td></tr>", mailBody, record.days)
                } else if record.days == 90 {
                        mailBody = fmt.Sprintf("%s<td align=\"center\"style=\"background-color: #8BE78B\">%d</td></tr>", mailBody, record.days)
                }
        }
        return mailBody
}

func sendMail(body string) {
        msg := gomail.NewMessage()
        msg.SetHeader("From", "racktables@dedale.tf1.fr")
        msg.SetHeader("To", "asantangeli@tf1.fr")
        msg.SetHeader("Subject", "RackTables Expirations Report")
        msg.SetBody("text/html", body)
        mailer := gomail.NewMailer("prdinfmel501-adm.dedale.tf1.fr", "", "", 25)
        if err := mailer.Send(msg); err != nil {
                panic(err)
        }
}

func main() {
        httpGetWithAuth()
        mapRequestParsing()
        for _, object := range expirations {
                if object.days == 30 || object.days == 60 || object.days == 90 {
                        body := tableauHTML()
                        sendMail(body)
                        return
                }
        }
}
