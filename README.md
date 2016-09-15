# edt2ical

Cet utilitaire permet de convertir l'emploi du temps fourni par le secrétariat du M1 Informatique de l'Université Paris-Sud dans le format ICAL.

ICAL est utilisé par Google Calendar et la plupart des applis d'agenda.

## Installation

* https://golang.org/doc/install
* Cloner le dépôt puis `cd edt2ical`
* `go get`
* `go build`

## Utilisation

Il faut tout d'abord convertir l'emploi du temps au format PDF en XLSX (Microsoft Excel) ou ODS (LibreOffice Calc), par exemple via ce site :

https://online2pdf.com/pdf2excel

L'ouvrir ensuite à l'aide d'un tableur et le convertir en CSV.

On pourra ainsi le transformer encore une fois :
```
./edt2ical -file fichier.csv >edt.ical
```
