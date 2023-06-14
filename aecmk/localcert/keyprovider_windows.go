package localcert

import (
	"crypto/x509"
	"fmt"
	"strings"
	"unsafe"

	mssql "github.com/microsoft/go-mssqldb"

	"github.com/microsoft/go-mssqldb/internal/certs"
	"golang.org/x/sys/windows"
)

var WindowsCertificateStoreKeyProvider = LocalCertProvider{name: mssql.CertificateStoreKeyProvider, passwords: make(map[string]string)}

func init() {
	mssql.RegisterCekProvider(mssql.CertificateStoreKeyProvider, &WindowsCertificateStoreKeyProvider)
}

func (p *LocalCertProvider) loadWindowsCertStoreCertificate(path string) (privateKey interface{}, cert *x509.Certificate) {
	privateKey = nil
	cert = nil
	pathParts := strings.Split(path, `/`)
	if len(pathParts) != 3 {
		panic(invalidCertificatePath(path, fmt.Errorf("key store path requires 3 segments")))
	}

	var storeId uint32
	switch strings.ToLower(pathParts[0]) {
	case "localmachine":
		storeId = windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
	case "currentuser":
		storeId = windows.CERT_SYSTEM_STORE_CURRENT_USER
	default:
		panic(invalidCertificatePath(path, fmt.Errorf("Unknown certificate store")))
	}
	system, err := windows.UTF16PtrFromString(pathParts[1])
	if err != nil {
		panic(err)
	}
	h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM,
		windows.PKCS_7_ASN_ENCODING|windows.X509_ASN_ENCODING,
		0,
		storeId, uintptr(unsafe.Pointer(system)))
	if err != nil {
		panic(err)
	}
	defer windows.CertCloseStore(h, 0)
	signature := thumbprintToByteArray(pathParts[2])
	return certs.FindCertBySignatureHash(h, signature)
}
