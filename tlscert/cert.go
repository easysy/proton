package tlscert

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"
	"time"
)

var (
	FSIsEmpty           = errors.New("the embedded file system is empty")
	CertFilePathIsEmpty = errors.New("the path to the certificate file is empty")
	KeyFilePathIsEmpty  = errors.New("the path to the key file is empty")
	CertPEMBlockIsEmpty = errors.New("PEM certificate block is empty")
	KeyPEMBlockIsEmpty  = errors.New("PEM key block is empty")
	NoValid             = errors.New("no valid certificate")
	AppendCertFailed    = errors.New("failed to add CA's certificate")
)

type CertificatesLoader func() ([]tls.Certificate, *x509.CertPool, error)

func ServerTLSConfig(loader CertificatesLoader) (*tls.Config, error) {
	certificates, certPool, err := loader()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: certificates,
		ClientCAs:    certPool,
	}, nil
}

func ClientTLSConfig(loader CertificatesLoader) (*tls.Config, error) {
	certificates, certPool, err := loader()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: certificates,
		RootCAs:      certPool,
	}, nil
}

type Loader struct {
	EmbedFS                   *embed.FS
	CertFilePath, KeyFilePath string
	CertPEMBlock, KeyPEMBlock []byte

	Template    *x509.Certificate
	LoadRootCAs CertificatesLoader
	Bits        int
}

func certValidate(certificate *tls.Certificate) error {
	now := time.Now()

	for _, cert := range certificate.Certificate {
		certTemp, err := x509.ParseCertificate(cert)
		if err != nil {
			continue
		}

		if certTemp.NotBefore.Before(now) && certTemp.NotAfter.After(now) {
			return nil
		}
	}

	return NoValid
}

// LoadFromEmbed loads []tls.Certificate from a pair of files stored in *embed.FS.
// To use this function you must specify EmbedFS, CertFilePath, KeyFilePath in the Loader.
func (l *Loader) LoadFromEmbed() ([]tls.Certificate, *x509.CertPool, error) {
	if l.EmbedFS == nil {
		return nil, nil, FSIsEmpty
	}

	if l.CertFilePath == "" {
		return nil, nil, CertFilePathIsEmpty
	}

	if l.KeyFilePath == "" {
		return nil, nil, KeyFilePathIsEmpty
	}

	var err error

	var certPEMBlock []byte
	if certPEMBlock, err = l.EmbedFS.ReadFile(l.CertFilePath); err != nil {
		return nil, nil, err
	}

	var keyPEMBlock []byte
	if keyPEMBlock, err = l.EmbedFS.ReadFile(l.KeyFilePath); err != nil {
		return nil, nil, err
	}

	var certificate tls.Certificate
	if certificate, err = tls.X509KeyPair(certPEMBlock, keyPEMBlock); err != nil {
		return nil, nil, err
	}

	if err = certValidate(&certificate); err != nil {
		return nil, nil, err
	}

	return []tls.Certificate{certificate}, nil, nil
}

// LoadFromFiles loads []tls.Certificate from a pair of files.
// To use this function you must specify CertFilePath, KeyFilePath in the Loader.
func (l *Loader) LoadFromFiles() ([]tls.Certificate, *x509.CertPool, error) {
	if l.CertFilePath == "" {
		return nil, nil, CertFilePathIsEmpty
	}

	if l.KeyFilePath == "" {
		return nil, nil, KeyFilePathIsEmpty
	}

	certificate, err := tls.LoadX509KeyPair(l.CertFilePath, l.KeyFilePath)
	if err != nil {
		return nil, nil, err
	}

	if err = certValidate(&certificate); err != nil {
		return nil, nil, err
	}

	return []tls.Certificate{certificate}, nil, nil
}

// LoadFromBytes loads []tls.Certificate from a pair of PEM encoded data.
// To use this function you must specify CertPEMBlock, KeyPEMBlock in the Loader.
func (l *Loader) LoadFromBytes() ([]tls.Certificate, *x509.CertPool, error) {
	if len(l.CertPEMBlock) == 0 {
		return nil, nil, CertPEMBlockIsEmpty
	}

	if len(l.KeyPEMBlock) == 0 {
		return nil, nil, KeyPEMBlockIsEmpty
	}

	certificate, err := tls.X509KeyPair(l.CertPEMBlock, l.KeyPEMBlock)
	if err != nil {
		return nil, nil, err
	}

	if err = certValidate(&certificate); err != nil {
		return nil, nil, err
	}

	return []tls.Certificate{certificate}, nil, nil
}
