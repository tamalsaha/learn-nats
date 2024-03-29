package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

var cert_b64 = `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJYWY3NXZwZW91ZGd3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRBMk1qUXhPREUzTkROYUZ3MHlNakEyTWpReE9ERTNORFphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXpYdUZNQ211OHZqekNUcDQKYndDYjdCc3p3Zi9Tcy8xWGtjTzNWNC9haGxLbUZEeUIwbDJVMmtqTFFDNEdKQ1lpemVXcGFwcUJkdm9wSnVnQgpMc29jVzlreHlFYkVxdVVPU2IvQkYvS0FiNlhBdm14L0huQnNVSjJ4YTNjdDRTcGF5RTZJK0o0QkpRUi9vYUpxClJBcDltQm1YdG5sMmMrbmNpcWZRNks0cUtlbS82enFad2JXZFlEQzN4TVJRajBFcXQvb1FLdDdnWUdxVW9kTUUKU3hqVFFBVUR5K2dGVHNYUkdHS21QYmJJTXg3bXJ2KzgyMzEra0hvWlo3aXBCM1ViWDVpaVBwdTNmS0o5c3J0YQo3ZXNQNHlvWGZiUVhhWG1IdTNUc2hOeHlRMzN3MG5SUGxlQm5mT21HM3FFSkgyb2o0UkY5U1VNNHRpYmlheXFzClRic1l2d0lEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JRS2ZGVkNzYjhJZEdjczhjYzFPQlhhV1BzNwoxVEFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBaGJ5bzU3NHliZGdEeTkvK3BpVm1NYytkc0s2WU9RUzJOUkhOCnppQXhxbU9mcm04ZFhudnRIaEw0YTdtVXBxa1UyZnVCcXNqQzFqdC9pbUR2RGF3OHhtbmNPZFd6VWM2ckhJYncKbk1jeDdIYmNKRzZsZFFrT0NUc3NXZTRHYXNHa3A1TkZIcUdsaTZaMmNDeHNkdlFhczlFZiswVkl4c281RUprbwpVOTZpZFZJYnF3MWJxeVFocmlxZWk1dXcxcE9hTEw4YjBYTERMTnBacTM0SWJFVyt2ZlB0ZkVSMENLTmdMeGwvCjVrQUlaamVpNDRYL3JpOG5jTUxqdmVQOWNhWjZHbGtqcjEyNTFsV2VaRWlQQkVQM1hJOStTRjhPRTBPZys3VXUKdVJsVkRZNjZvZzVCNnFZRXI1QnJweHcva0xGZlhhdzFrcU82Zy9oOFUxK0xDSTJTYkE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==`

var key_b64 = `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBelh1Rk1DbXU4dmp6Q1RwNGJ3Q2I3QnN6d2YvU3MvMVhrY08zVjQvYWhsS21GRHlCCjBsMlUya2pMUUM0R0pDWWl6ZVdwYXBxQmR2b3BKdWdCTHNvY1c5a3h5RWJFcXVVT1NiL0JGL0tBYjZYQXZteC8KSG5Cc1VKMnhhM2N0NFNwYXlFNkkrSjRCSlFSL29hSnFSQXA5bUJtWHRubDJjK25jaXFmUTZLNHFLZW0vNnpxWgp3YldkWURDM3hNUlFqMEVxdC9vUUt0N2dZR3FVb2RNRVN4alRRQVVEeStnRlRzWFJHR0ttUGJiSU14N21ydis4CjIzMStrSG9aWjdpcEIzVWJYNWlpUHB1M2ZLSjlzcnRhN2VzUDR5b1hmYlFYYVhtSHUzVHNoTnh5UTMzdzBuUlAKbGVCbmZPbUczcUVKSDJvajRSRjlTVU00dGliaWF5cXNUYnNZdndJREFRQUJBb0lCQVFDQ1NLd1U4b2p6azNiOQpSZTV3YXhGeHJYbXVxcGFjK3FlWVMyQ25DeFhDRHdzd1Q0RDhzY3NjY0FVMjV6ZUxtZ1o5Ui8yWUV1aTlXRFhaCmJrYTV0UG93SGxFYkxBdXNVMWt3MTMwRnd3TStSdmtqZzhWQnRvUm14T1ZtUHdWKys0emQ3aldZZFE1Q3UweDEKWG5aRU4rYVVGcjREdTVXb1B3SlBnOEhJbGcwenp0ZEgrMzV6bkhUcFBIaXlvS3pvUVF4WWIwd05RQ1JZUlE4awpQa0tYQTJxWEQwS3VYeGRmeVZTY3A3M2RaY0prVzkzM3Q2YjZTTWVoeEI0UmY2aWMrcFJYeCtHdmJpWnpLSUlICjcxaC9HZzBNZ1B6NWxEd2MvMGZXYzBtUmlZVHIzOVhlSWF3eDY0d0pmNnh4cUZxR2FqZFFWaTNkaGFGdUxlRGwKSFR4R1NZaGhBb0dCQU5pYjEvbGd5cUcvTXhCSUgxcHlmbWhDdHU0Y1JiRmFYV3VSaGFCdVRJcHh3WkhjQXFGbQpMSXF2eDdYMTRZRWh5b20wTUZ5eDVsaXFtK2tQWDhBRG42dEVQZWZwSlF4YWFGYTlvTHZUNk9kTGIvMlNnSkZ4ClpBRUw5SG10R05ySDJPWmJqR2Vna25ITkVjOUQrUlZhUXNZU1VnTU1mdzdQV0Q4U0w1OGtNTUVQQW9HQkFQTFoKc2w2Smt0bHdNM2dqOXFOWG84OW50RnAwc2E1NTBxTHBCbEh1eitqT0dxeW9jMWxoYWZuZmZQYUxsNHd3ODR6ZQp4OTRUMGRMMnhNQjhWeVhFMkM4RUJhVHQ1UzZKSUEzSUdtYzRad2poTFFXZ3NvU1didERSK0g5Nnp5cEJjRnJ4CmUrK1pvaUVBMUFiNHhSRXJ3ZWZoVlFMbmhiU3BvT014S3pJYSs4MVJBb0dCQUxLUkR5UERXbWlySWFKN2duVmkKeTdpUnZ4SmVka20xMEN2Y1pJZVVSajhmZGs4VFM0dllta0dlbFluNDhIVXU4VFJDT2xoQVJEKzJMaCtjai9mUQpSUEhBcVRRazdHalpvd2hXL1VtNmNWY3p4bGdKVFRvWmV6S3RzMVlYajlUVVNZZmwwc0tmQ2ZzTTduQ3FmWTNQCndocGRnZ1NIYWJ0QXpXUDVUdzdubTlXYkFvR0FBUmJrNi9PbUN2K3IyM0FkM1NHNWhHYXNzbk12a043UENSZ08KaFRPVER6Sk5nRlRKSDYrR01DN0dlcnlwazJGczFrYnhrcGQ0SzRBYjVka284dXh0STlqYXhhQ2psSS9jNnZMbwoyMm12WEtUVjlONkJyb0tXUUsyUWRkSHhOL2xQTGJsRG14R1BYcUtJVVBld3VxRDluN0t0RlBSQTcweUxnamxvClBqTk15ZUVDZ1lCeG9GZFpJRVFzNlNoVzRuNC93Mm1mcFB5eFNwMHdRanNLb0RDUnF6dmFuRTE1N1d0M1N2ZWYKWkVvTWdhV1l0L081dlM1bFgycVozVytXbjNKakdJajRQVnVsMSs0aWhVUDhHdmhaYnZTZVFPTjN1NG5PYW1veQorUkZTdmMrbVpwVDBWNWxnbE9vNlBlYTJ2emFneCtDWXFrRG0zMUZWWE4yVWkzNEkzZEpGamc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`

func main() {
	cert, err := tls.LoadX509KeyPair("/tmp/k2.key", "/tmp/k2.key")
	if err != nil {
		panic(err)
	}
	fmt.Println(cert.Certificate)
}

func main__() {
	cert_bytes, err := base64.StdEncoding.DecodeString(cert_b64)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/kind.crt", cert_bytes, 0o644)
	if err != nil {
		panic(err)
	}

	key_bytes, err := base64.StdEncoding.DecodeString(key_b64)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/kind.key", key_bytes, 0o644)
	if err != nil {
		panic(err)
	}

	cert, err := tls.LoadX509KeyPair("/tmp/kind.crt", "/tmp/kind.key")
	if err != nil {
		panic(err)
	}

	// tls.X509KeyPair()

	var buf bytes.Buffer

	pemCrt := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   cert.Certificate[0],
	}
	err = pem.Encode(&buf, pemCrt)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/k2.crt", buf.Bytes(), 0o644)
	if err != nil {
		panic(err)
	}

	// https://golang.org/src/crypto/tls/tls.go?s=7880:7947#L370
	k2, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		panic(err)
	}
	pemKey := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   k2,
	}
	// buf.Reset() // ------------------
	err = pem.Encode(&buf, pemKey)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/k2.key", buf.Bytes(), 0o644)
	if err != nil {
		panic(err)
	}

	c2, err := tls.LoadX509KeyPair("/tmp/k2.crt", "/tmp/k2.key")
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(c2)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/k2.json", data, 0o644)
	if err != nil {
		panic(err)
	}

	data, err = json.Marshal(cert)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("/tmp/kind_cert.json", data, 0o644)
	if err != nil {
		panic(err)
	}
}

func ToX509CombinedKeyPair(cert *tls.Certificate) ([]byte, error) {
	var buf bytes.Buffer

	for _, c := range cert.Certificate {
		if err := pem.Encode(&buf, &pem.Block{
			Type:    "CERTIFICATE",
			Headers: nil,
			Bytes:   c,
		}); err != nil {
			return nil, err
		}
	}

	// https://golang.org/src/crypto/tls/tls.go?s=7880:7947#L370
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   key,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ToX509KeyPair(cert *tls.Certificate) (certPEMBlock, keyPEMBlock []byte, err error) {
	var buf bytes.Buffer

	for _, c := range cert.Certificate {
		if err := pem.Encode(&buf, &pem.Block{
			Type:    "CERTIFICATE",
			Headers: nil,
			Bytes:   c,
		}); err != nil {
			return nil, nil, err
		}
	}
	certPEMBlock = buf.Bytes()
	buf.Reset()

	// https://golang.org/src/crypto/tls/tls.go?s=7880:7947#L370
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return nil, nil, err
	}
	err = pem.Encode(&buf, &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   key,
	})
	if err != nil {
		return nil, nil, err
	}
	keyPEMBlock = buf.Bytes()

	return
}
