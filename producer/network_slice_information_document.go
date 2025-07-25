// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0
//

/*
 * NSSF NS Selection
 *
 * NSSF Network Slice Selection Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package producer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/omec-project/nssf/logger"
	stats "github.com/omec-project/nssf/metrics"
	"github.com/omec-project/nssf/plugin"
	"github.com/omec-project/nssf/util"
	"github.com/omec-project/openapi/models"
	"github.com/omec-project/util/httpwrapper"
)

// Parse NSSelectionGet query parameter
func parseQueryParameter(query url.Values) (plugin.NsselectionQueryParameter, error) {
	var (
		param plugin.NsselectionQueryParameter
		err   error
	)

	if query.Get("nf-type") != "" {
		param.NfType = new(models.NfType)
		*param.NfType = models.NfType(query.Get("nf-type"))
	}

	param.NfId = query.Get("nf-id")

	if query.Get("slice-info-request-for-registration") != "" {
		param.SliceInfoRequestForRegistration = new(models.SliceInfoForRegistration)
		err = json.NewDecoder(strings.NewReader(
			query.Get("slice-info-request-for-registration"))).Decode(param.SliceInfoRequestForRegistration)
		if err != nil {
			return param, err
		}
	}

	if query.Get("slice-info-request-for-pdu-session") != "" {
		param.SliceInfoRequestForPduSession = new(models.SliceInfoForPduSession)
		err = json.NewDecoder(strings.NewReader(
			query.Get("slice-info-request-for-pdu-session"))).Decode(param.SliceInfoRequestForPduSession)
		if err != nil {
			return param, err
		}
	}

	if query.Get("home-plmn-id") != "" {
		param.HomePlmnId = new(models.PlmnId)
		err = json.NewDecoder(strings.NewReader(query.Get("home-plmn-id"))).Decode(param.HomePlmnId)
		if err != nil {
			return param, err
		}
	}

	if query.Get("tai") != "" {
		param.Tai = new(models.Tai)
		err = json.NewDecoder(strings.NewReader(query.Get("tai"))).Decode(param.Tai)
		if err != nil {
			return param, err
		}
	}

	if query.Get("supported-features") != "" {
		param.SupportedFeatures = query.Get("supported-features")
	}

	return param, err
}

// Check if the NF service consumer is authorized
// TODO: Check if the NF service consumer is legal with local configuration, or possibly after querying NRF through `nf-id` e.g. Whether the V-NSSF is authorized
func checkNfServiceConsumer(nfType models.NfType) error {
	if nfType != models.NfType_AMF && nfType != models.NfType_NSSF {
		return fmt.Errorf("`nf-type`:'%s' is not authorized to retrieve the slice selection information", string(nfType))
	}

	return nil
}

// NSSelectionGet - Retrieve the Network Slice Selection Information
func HandleNSSelectionGet(request *httpwrapper.Request) *httpwrapper.Response {
	logger.Nsselection.Infof("Handle NSSelectionGet")

	query := request.Query

	response, problemDetails := NSSelectionGetProcedure(query)

	nfType := GetNfTypeFromQueryParameters(query)
	nfId := GetNfIdFromQueryParameters(query)

	if response != nil {
		stats.IncrementNssfNsSelectionsStats(nfType, nfId, "SUCCESS")
		return httpwrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		stats.IncrementNssfNsSelectionsStats(nfType, nfId, "FAILURE")
		return httpwrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	stats.IncrementNssfNsSelectionsStats(nfType, nfId, "FAILURE")
	return httpwrapper.NewResponse(http.StatusForbidden, nil, problemDetails)
}

func NSSelectionGetProcedure(query url.Values) (*models.AuthorizedNetworkSliceInfo, *models.ProblemDetails) {
	var (
		status         int
		response       *models.AuthorizedNetworkSliceInfo
		problemDetails *models.ProblemDetails
	)
	response = &models.AuthorizedNetworkSliceInfo{}
	problemDetails = &models.ProblemDetails{}

	// TODO: Record request times of the NF service consumer and response with ProblemDetails of 429 Too Many Requests
	//       if the consumer has sent too many requests in a configured amount of time
	// TODO: Check URI length and response with ProblemDetails of 414 URI Too Long if URI is too long

	// Parse query parameter
	param, err := parseQueryParameter(query)
	if err != nil {
		// status = http.StatusBadRequest
		problemDetails = &models.ProblemDetails{
			Title:  util.MALFORMED_REQUEST,
			Status: http.StatusBadRequest,
			Detail: "[Query Parameter] " + err.Error(),
		}
		return nil, problemDetails
	}

	// Check permission of NF service consumer
	err = checkNfServiceConsumer(*param.NfType)
	if err != nil {
		// status = http.StatusForbidden
		problemDetails = &models.ProblemDetails{
			Title:  util.UNAUTHORIZED_CONSUMER,
			Status: http.StatusForbidden,
			Detail: err.Error(),
		}
		return nil, problemDetails
	}

	if param.SliceInfoRequestForRegistration != nil {
		// Network slice information is requested during the Registration procedure
		status = nsselectionForRegistration(param, response, problemDetails)
	} else {
		// Network slice information is requested during the PDU session establishment procedure
		status = nsselectionForPduSession(param, response, problemDetails)
	}

	if status != http.StatusOK {
		return nil, problemDetails
	}
	return response, nil
}

func GetNfTypeFromQueryParameters(query url.Values) (nfType string) {
	if query.Get("nf-type") != "" {
		return query.Get("nf-type")
	}
	return "UNKNOWN_NF"
}

func GetNfIdFromQueryParameters(query url.Values) (nfId string) {
	if query.Get("nf-id") != "" {
		return query.Get("nf-id")
	}
	return "UNKNOWN_NF_ID"
}
