/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"encoding/json"
	"github.com/apache/incubator-devlake/config"
	"github.com/apache/incubator-devlake/logger"
	"github.com/apache/incubator-devlake/mocks"
	"github.com/apache/incubator-devlake/models/common"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/gitlab/tasks"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/runner"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessScope(t *testing.T) {
	cfg := config.GetConfig()
	log := logger.Global.Nested("gitlab")
	db, _ := runner.NewGormDb(cfg, log)
	Init(cfg, log, db)
	connection := &models.GitlabConnection{
		RestConnection: helper.RestConnection{
			BaseConnection: helper.BaseConnection{
				Name: "gitlab-test",
				Model: common.Model{
					ID: 1,
				},
			},
			Endpoint:         "https://gitlab.com/api/v4/",
			Proxy:            "",
			RateLimitPerHour: 0,
		},
		AccessToken: helper.AccessToken{
			Token: "123",
		},
	}
	mockMeta := mocks.NewPluginMeta(t)
	mockMeta.On("RootPkgPath").Return("github.com/apache/incubator-devlake/plugins/gitlab")
	err := core.RegisterPlugin("gitlab", mockMeta)
	assert.Nil(t, err)
	bs := &core.BlueprintScopeV100{
		Entities: []string{"CODE"},
		Options: json.RawMessage(`{
              "projectId": 123
            }`),
		Transformation: json.RawMessage(`{
              "prType": "hey,man,wasup",
              "refdiff": {
                "tagsPattern": "pattern",
                "tagsLimit": 10,
                "tagsOrder": "reverse semver"
              },
              "dora": {
                "environment": "pattern",
                "environmentRegex": "xxxx"
              }
            }`),
	}
	apiRepo := &tasks.GitlabApiProject{
		GitlabId:      123,
		HttpUrlToRepo: "HttpUrlToRepo",
	}
	scopes := make([]*core.BlueprintScopeV100, 0)
	scopes = append(scopes, bs)
	plan := make(core.PipelinePlan, len(scopes))
	for i, scopeElem := range scopes {
		plan, err = processScope(nil, 1, scopeElem, i, plan, apiRepo, connection)
		assert.Nil(t, err)
	}
	planJson, err1 := json.Marshal(plan)
	assert.Nil(t, err1)
	expectPlan := `[[{"plugin":"gitlab","subtasks":[],"options":{"connectionId":1,"projectId":123,"transformationRules":{"prType":"hey,man,wasup"}}},{"plugin":"gitextractor","subtasks":null,"options":{"proxy":"","repoId":"gitlab:GitlabProject:1:123","url":"//git:123@HttpUrlToRepo"}}],[{"plugin":"refdiff","subtasks":null,"options":{"tagsLimit":10,"tagsOrder":"reverse semver","tagsPattern":"pattern"}}],[{"plugin":"dora","subtasks":null,"options":{"repoId":"gitlab:GitlabProject:1:123","tasks":["EnrichTaskEnv"],"transformation":{"environment":"pattern","environmentRegex":"xxxx"}}}]]`
	assert.Equal(t, expectPlan, string(planJson))
}
