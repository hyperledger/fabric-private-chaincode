package utils

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func MarshalProtoBase64(msg proto.Message) string {
	bytes, _ := proto.Marshal(msg)
	return base64.StdEncoding.EncodeToString(bytes)
}

func UnmarshalProtoBase64(raw []byte, msg proto.Message) error {
	bytes, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return errors.Wrap(err, "cannot base64 decode")
	}

	if err := proto.Unmarshal(bytes, msg); err != nil {
		return errors.Wrap(err, "cannot unmarshal proto")
	}

	return nil
}
