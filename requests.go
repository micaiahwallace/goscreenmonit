package goscreenmonit

import (
	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Create a client request container with a message
func CreateRequest(reqtype uploadpb.ClientRequest_RequestType, message protoreflect.ProtoMessage) ([]byte, error) {

	// Get the request bytes
	reqbytes, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	// Create a request container
	request := &uploadpb.ClientRequest{
		Type:    reqtype,
		Request: reqbytes,
	}

	// Serialize data
	return proto.Marshal(request)
}

// Create a registration message
func CreateRegistration(host, user string) ([]byte, error) {

	// Create registration command
	regcmd := &uploadpb.Register{
		Host: host,
		User: user,
	}

	return CreateRequest(uploadpb.ClientRequest_REGISTER, regcmd)
}

// Create an image upload message
func CreateUpload(images [][]byte) ([]byte, error) {

	// create the message
	msg := &uploadpb.ImageUpload{
		Images:    images,
		Timestamp: timestamppb.Now(),
	}

	// Serialize data
	return CreateRequest(uploadpb.ClientRequest_UPLOAD, msg)
}
