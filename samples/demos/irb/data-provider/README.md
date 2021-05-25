# Data provider

## Testing

```
make
```

## Web service

The webservice received patients consents as PDFs. To parse those files we use `pdftotext`. 

Please install:
```
apt install poppler-utils
```

You can test the parsing manually as follows:
```
cd service/pdf
go run cmd/main.go -input=patient1.pdf
go run cmd/main.go -input=patient2.pdf
```

To run the data-provider web service
```
cd service
go run .
```

By default, it starts listing on port 3000. To change the port use `go run . -port=4000`

### Dummy patients

The web service generates on startup two dummy patients (`patient1` and `patient2`).