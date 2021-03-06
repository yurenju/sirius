// Copyright 2018 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package health

import (
	"context"

	"github.com/getamis/sirius/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CheckFn func(context.Context) error

// defaultServer is the implementation of HealthCheckServiceServer
type defaultServer struct {
	checkFns []CheckFn
}

func New(checkFns ...CheckFn) HealthCheckServiceServer {
	return &defaultServer{
		checkFns: checkFns,
	}
}

// Liveness is represented that whether application is able to make progress or not.
func (s *defaultServer) Liveness(ctx context.Context, req *EmptyRequest) (*EmptyResponse, error) {
	return nil, nil
}

// Readiness is represented that whether application is ready to start accepting traffic or not.
func (s *defaultServer) Readiness(ctx context.Context, req *EmptyRequest) (*EmptyResponse, error) {
	if len(s.checkFns) == 0 {
		return nil, nil
	}
	errCh := make(chan error, len(s.checkFns))
	for _, checker := range s.checkFns {
		go func(checker CheckFn) {
			errCh <- checker(ctx)
		}(checker)
	}
	for _ = range s.checkFns {
		if err := <-errCh; err != nil {
			log.Error("Failed to check readiness", "err", err)
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}
	return nil, nil
}
