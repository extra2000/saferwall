// Copyright 2020 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"context"

	"github.com/saferwall/saferwall/pkg/grpc/multiav"
	pb "github.com/saferwall/saferwall/pkg/grpc/multiav/sophos/proto"
	"github.com/saferwall/saferwall/pkg/multiav/sophos"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

type server struct {
	avDbUpdateDate int64
}

// ScanGetVersionFile implements sophos.SophosScanner.
func (s *server) GetVersion(ctx context.Context, in *pb.VersionRequest) (*pb.VersionResponse, error) {
	version, err := sophos.GetVersion()
	return &pb.VersionResponse{Version: version.ProductVersion}, err
}

// ScanFile implements sophos.SophosScanner.
func (s *server) ScanFile(ctx context.Context, in *pb.ScanFileRequest) (*pb.ScanResponse, error) {
	res, err := sophos.ScanFile(in.Filepath)
	return &pb.ScanResponse{
		Infected: res.Infected,
		Output:   res.Output,
		Update:   s.avDbUpdateDate}, err
}

// main start a gRPC server and waits for connection.
func main() {
	args := os.Args[1:]
	port := args[0]

	// create a listener on TCP port
	lis, err := multiav.CreateListener(":"+port)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	// create a gRPC server object
	s := multiav.NewServer()

	// get the av db update date
	avDbUpdateDate, err := multiav.UpdateDate()
	if err != nil {
		grpclog.Fatalf("failed to read av db update date %v", err)
	}

	// attach the SophosScanner service to the server
	pb.RegisterSophosScannerServer(
		s, &server{avDbUpdateDate: avDbUpdateDate})

	// register reflection service on gRPC server and serve.
	log.Infoln("Starting Sophos gRPC server ...")
	err = multiav.Serve(s, lis)
	if err != nil {
		grpclog.Fatalf("failed to serve: %v", err)
	}
}
