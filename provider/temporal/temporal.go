package temporal

import (
	"github.com/upper-institute/graviflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

type TemporalBuilder[Dependency any] struct {
	*graviflow.AppInjector[Dependency]

	Namespace graviflow.Config `config:"namespace,string" default:"default" usage:"Temporal namespace"`
	HostPort  graviflow.Config `config:"address,string" default:"localhost:7233" usage:"Tempora server host:port"`
	// AesCodecKey    graviflow.Config `config:"aes.codec.key,string" usage:"Temporal AES codec key"`
	// AesCodecIvSeed graviflow.Config `config:"aes.iv.seed,int64" usage:"Temporal AES codec IV rand seed"`
}

func (t *TemporalBuilder[Dependency]) GetTemporalClient() client.Client {

	log := t.Log()

	ns := t.Namespace.StringVal()
	if len(ns) == 0 {
		ns = "default"
	}

	dataConv := converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewProtoJSONPayloadConverter(),
		converter.NewJSONPayloadConverter(),
	)

	// if t.AesCodecKey.IsSet() {

	// 	block, err := aes.NewCipher([]byte(t.AesCodecKey.StringVal()))
	// 	if err != nil {
	// 		log.Panic("Error while creating aes cipher for Temporal Codec Converter", "error", err)
	// 	}

	// 	dataConv = converter.NewCodecDataConverter(
	// 		dataConv,
	// 		&cipherTemporalCodec{
	// 			block: block,
	// 			rand:  rand.New(rand.NewSource(t.AesCodecIvSeed.Int64Val())),
	// 		},
	// 	)

	// }

	opts := client.Options{
		Namespace:     ns,
		HostPort:      t.HostPort.StringVal(),
		Logger:        t.Log(),
		DataConverter: dataConv,
	}

	cli, err := client.Dial(opts)
	if err != nil {
		log.Panic("Unable to dial temporal server", "error", err)
	}

	return cli

}

// type cipherTemporalCodec struct {
// 	block cipher.Block
// 	rand  *rand.Rand
// }

// func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
// 	padding := (blockSize - len(ciphertext)%blockSize)
// 	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
// 	return append(ciphertext, padtext...)
// }

// func (c *cipherTemporalCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {

// 	iv := make([]byte, 16)

// 	_, err := c.rand.Read(iv)
// 	if err != nil {
// 		return nil, err
// 	}

// 	mode := cipher.NewCBCEncrypter(c.block, iv)

// 	result := make([]*commonpb.Payload, len(payloads))
// 	for i, p := range payloads {

// 		raw, err := proto.Marshal(p)
// 		if err != nil {
// 			return payloads, err
// 		}

// 		raw = graviflow.PKCS5Padding(raw, mode.BlockSize())

// 		encrypted := make([]byte, len(raw))

// 		mode.CryptBlocks(encrypted, raw)

// 		result[i] = &commonpb.Payload{
// 			Metadata: map[string][]byte{
// 				"MetadataEncoding": []byte("binary/aes"),
// 				"IV":               iv,
// 			},
// 			Data: encrypted,
// 		}

// 	}

// 	return result, nil

// }

// func (c *cipherTemporalCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {

// 	result := make([]*commonpb.Payload, len(payloads))
// 	for i, p := range payloads {

// 		if string(p.Metadata["MetadataEncoding"]) != "binary/aes" {
// 			result[i] = p
// 			continue
// 		}

// 		mode := cipher.NewCBCDecrypter(c.block, p.Metadata["IV"])

// 		decrypted := make([]byte, len(p.Data))

// 		mode.CryptBlocks(decrypted, p.Data)

// 		raw := graviflow.PKCS5Trimming(decrypted)

// 		result[i] = &commonpb.Payload{}
// 		err := proto.Unmarshal(raw, result[i])
// 		if err != nil {
// 			return payloads, err
// 		}

// 	}

// 	return result, nil

// }
