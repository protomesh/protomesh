package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"time"

	"github.com/upper-institute/graviflow"
)

type KeyLoader[Dependency any] struct {
	*graviflow.AppInjector[Dependency]

	KeysPath graviflow.Config `config:"path,str" usage:"Path to PEM encoded private key file"`

	priv crypto.PrivateKey
}

func (k *KeyLoader[Dependency]) BuildPrivateKeys() []crypto.PrivateKey {

	if !k.KeysPath.IsSet() {
		return nil
	}

	log := k.Log()

	pemData, err := ioutil.ReadFile(k.KeysPath.StringVal())
	if err != nil {
		log.Panic("Failed to read private keys file", "error", err, "keysPath", k.KeysPath.StringVal())
	}

	privs := []crypto.PrivateKey{}

	for {
		var block *pem.Block

		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		switch block.Type {

		case "EC PRIVATE KEY":
			priv, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				log.Panic("Unable to parse EC (SEC) private key", "error", err)
			}
			privs = append(privs, priv)

		case "RSA PRIVATE KEY":
			priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				log.Panic("Unable to parse PKCS1 private key", "error", err)
			}
			privs = append(privs, priv)

		case "PRIVATE KEY":
			priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				log.Panic("Unable to parse PKCS8 private key", "error", err)
			}
			privs = append(privs, priv)

		}
	}

	return privs

}

func (k *KeyLoader[Dependency]) BuildDefaultRSAPrivateKey() []crypto.PrivateKey {

	privs := k.BuildPrivateKeys()
	if privs != nil && len(privs) > 0 {
		return privs
	}

	if k.priv != nil {
		return []crypto.PrivateKey{k.priv}
	}

	log := k.Log()

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Panic("Unable to generate temporary RSA private key", "error", err)
	}

	k.priv = priv

	return []crypto.PrivateKey{priv}

}

type CertificateLoader[Dependency any] struct {
	*graviflow.AppInjector[Dependency]

	CertificatePath graviflow.Config `config:"path" usage:"Path to PEM encoded certificate chain file"`
	PrivateKey      *KeyLoader[any]  `config:"private.key,str" usage:"Private key to sign default certificate"`
}

func (c *CertificateLoader[Dependency]) BuildCertificates() []*x509.Certificate {

	if !c.CertificatePath.IsSet() {
		return nil
	}

	log := c.Log()

	pemData, err := ioutil.ReadFile(c.CertificatePath.StringVal())
	if err != nil {
		log.Panic("Failed to read certificate file", "error", err, "privateKey", c.CertificatePath.StringVal())
	}

	certs := []*x509.Certificate{}

	for {

		var block *pem.Block

		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		switch block.Type {

		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Panic("Unable to parse certificate", "error", err)
			}

			certs = append(certs, cert)

		}

	}

	return certs

}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

func (c *CertificateLoader[Dependency]) BuildDefaultCertificate() []*x509.Certificate {

	log := c.Log()

	certs := c.BuildCertificates()
	if certs != nil && len(certs) > 0 {
		return certs
	}

	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:       []string{"Graviflow"},
			Country:            []string{},
			Province:           []string{},
			Locality:           []string{"Global"},
			OrganizationalUnit: []string{"Graviflow Agent"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(360 * 24 * time.Hour),
		IsCA:      true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
	}

	privKeys := c.PrivateKey.BuildDefaultRSAPrivateKey()

	privKey := privKeys[0]

	certBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, publicKey(privKey), privKey)
	if err != nil {
		log.Panic("Error self signing certificate", "error", err)
	}

	selfSignedCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		log.Panic("Error parsing self signed certificate", "error", err)
	}

	log.Info("Default RSA certificate created")

	return []*x509.Certificate{selfSignedCert}

}

type TlsCertificateLoader[Dependency any] struct {
	*graviflow.AppInjector[Dependency]
	Certificates *CertificateLoader[Dependency] `config:"certificates"`
}

func (t *TlsCertificateLoader[Dependency]) Build() *tls.Certificate {

	privs := t.Certificates.PrivateKey.BuildDefaultRSAPrivateKey()
	certs := t.Certificates.BuildDefaultCertificate()

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{},
		PrivateKey:  privs[0],
	}

	for _, cert := range certs {

		if tlsCert.Leaf == nil {
			tlsCert.Leaf = cert
		}

		tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
	}

	return tlsCert

}

type TlsBuilder[Dependency any] struct {
	*graviflow.AppInjector[Dependency]

	Certificate *TlsCertificateLoader[Dependency] `config:"certificate"`
	RootCAs     *CertificateLoader[Dependency]    `config:"root.cas"`

	InsecureSkipVerify graviflow.Config `config:"insecure.skip.verify,bool" default:"false" usage:"Skip server name verification against certificate chain"`

	ListenerAddress graviflow.Config `config:"listener.address,string" usage:"TLS listener address"`
	Protocol        graviflow.Config `config:"protocol,string" default:"tcp" usage:"Protocol to accept in the TLS listener"`
}

func (t *TlsBuilder[Dependency]) BuildConfig() *tls.Config {

	cert := t.Certificate.Build()

	tlsCfg := &tls.Config{
		Certificates:       []tls.Certificate{*cert},
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: t.InsecureSkipVerify.BoolVal(),
		NextProtos:         []string{"h2"},
	}

	if t.RootCAs != nil && t.RootCAs.CertificatePath.IsSet() {

		rootCAs := t.RootCAs.BuildCertificates()

		tlsCfg.RootCAs = x509.NewCertPool()

		for _, rootCA := range rootCAs {
			tlsCfg.RootCAs.AddCert(rootCA)
		}

	}

	return tlsCfg

}

func (t *TlsBuilder[Dependency]) BuildListener() net.Listener {

	log := t.Log()

	if t.ListenerAddress == nil || !t.ListenerAddress.IsSet() {
		log.Panic("ListenerAddress must be set in TlsBuilder")
	}

	if t.Protocol == nil || !t.Protocol.IsSet() {
		log.Panic("Protocol must be set in TlsBuilder")
	}

	proto := t.Protocol.StringVal()
	addr := t.ListenerAddress.StringVal()
	tlsCfg := t.BuildConfig()

	lis, err := tls.Listen(proto, addr, tlsCfg)
	if err != nil {
		log.Panic("Failed to listen (tls) to address", "error", err, "address", addr)
	}

	log.Info("Listening for TLS connections", "protocol", proto, "address", addr)

	return lis

}
