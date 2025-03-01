// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1alpha2

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"k8s.io/klog/v2"

	"strconv"
	"time"

	"kubesphere.io/kubesphere/pkg/api"
	meteringv1alpha1 "kubesphere.io/kubesphere/pkg/api/metering/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/metering"
	monitoringclient "kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func (h *tenantHandler) QueryMetering(req *restful.Request, resp *restful.Response) {

	u, ok := request.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, req, err)
		return
	}

	q := meteringv1alpha1.ParseQueryParameter(req)

	res, err := h.tenant.Metering(u, q, h.meteringOptions.Billing.PriceInfo)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	if q.Operation == monitoringv1alpha3.OperationExport {

		start, err := strconv.ParseInt(q.Start, 10, 64)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}
		end, err := strconv.ParseInt(q.End, 10, 64)
		if err != nil {
			api.HandleBadRequest(resp, nil, err)
			return
		}

		startTime := time.Unix(start, 0)
		endTime := time.Unix(end, 0)

		monitoringv1alpha3.ExportMetrics(resp, res, startTime, endTime)
		return
	}

	resp.WriteAsJson(res)
}

func (h *tenantHandler) QueryMeteringHierarchy(req *restful.Request, resp *restful.Response) {
	u, ok := request.UserFrom(req.Request.Context())
	if !ok {
		err := fmt.Errorf("cannot obtain user info")
		klog.Errorln(err)
		api.HandleForbidden(resp, req, err)
		return
	}

	q := meteringv1alpha1.ParseQueryParameter(req)
	q.Level = monitoringclient.LevelPod

	resourceStats, err := h.tenant.MeteringHierarchy(u, q, h.meteringOptions.Billing.PriceInfo)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}

	resp.WriteAsJson(resourceStats)
}

func (h *tenantHandler) HandlePriceInfoQuery(req *restful.Request, resp *restful.Response) {

	var priceResponse metering.PriceResponse

	priceInfo := h.meteringOptions.Billing.PriceInfo
	priceResponse.RetentionDay = h.meteringOptions.RetentionDay
	priceResponse.Currency = priceInfo.CurrencyUnit
	priceResponse.CpuPerCorePerHour = priceInfo.CpuPerCorePerHour
	priceResponse.MemPerGigabytesPerHour = priceInfo.MemPerGigabytesPerHour
	priceResponse.IngressNetworkTrafficPerMegabytesPerHour = priceInfo.IngressNetworkTrafficPerMegabytesPerHour
	priceResponse.EgressNetworkTrafficPerMegabytesPerHour = priceInfo.EgressNetworkTrafficPerMegabytesPerHour
	priceResponse.PvcPerGigabytesPerHour = priceInfo.PvcPerGigabytesPerHour

	resp.WriteAsJson(priceResponse)
}
