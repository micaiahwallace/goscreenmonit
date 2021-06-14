package main

import (
	"github.com/micaiahwallace/goscreenmonit/uploadpb"
	"google.golang.org/protobuf/proto"
)

// Create a server response container with a message
func CreateResponse(restype uploadpb.ServerResponse_MessageType) ([]byte, error) {

	// Create a request container
	request := &uploadpb.ServerResponse{
		Type: restype,
	}

	// Serialize data
	return proto.Marshal(request)
}
