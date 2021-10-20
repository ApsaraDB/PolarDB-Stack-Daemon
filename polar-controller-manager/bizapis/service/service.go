/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package service

import (
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"k8s.io/client-go/kubernetes"
)

// Service ...
type Service struct {
	System *SystemService
}

func NewService(client kubernetes.Interface, cfg *config.CompletedConfig) *Service {
	var (
		system = NewSystemService(cfg, client)
	)

	return &Service{
		System: system,
	}
}
