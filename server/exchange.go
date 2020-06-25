// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"crypto"
	"net/url"
	"path"

	"github.com/WICG/webpackage/go/signedexchange/version"
	"github.com/google/webpackager/certchain/certmanager"
	"github.com/google/webpackager/exchange"
)

// ExchangeMetaFactory is an exchange.FactoryProvider designed to be used
// with Handler.
type ExchangeMetaFactory struct {
	ExchangeConfig
}

var _ exchange.FactoryProvider = (*ExchangeMetaFactory)(nil)

// ExchangeConfig configures ExchangeMetaFactory.
type ExchangeConfig struct {
	// Version specifies the signed exchange version. If Version is empty,
	// ExchangeMetaFactory uses exchange.DefaultVersion.
	Version version.Version

	// MIRecordSize specifies Merkle Integrity record size. The value must
	// be positive, or zero to use exchange.DefaultMIRecordSize. It must not
	// exceed 16384 (16 KiB) to be compliant with the specification.
	MIRecordSize int

	// CertManager specifies an AugmentedChain provider. ExchangeMetaFactory
	// does not start or stop this CertManager automatically; the caller is
	// responsible to make the CertManager active before ExchangeMetaFactory
	// receives the first call of Get. CertManager may not be nil.
	CertManager *certmanager.Manager

	// CertURLBase specifies the base URL for the cert-url parameter in the
	// signature. ExchangeMetaFactory appends RawChain.Digest to CertURLBase,
	// as a stable unique identifier of the AugmentedChain, to construct the
	// cert-url parameter. CertURLBase may not be nil.
	CertURLBase *url.URL

	// PrivateKey specifies the private key used for signing. PrivateKey may
	// not be nil.
	PrivateKey crypto.PrivateKey

	// KeepNonSXGPreloads instructs Factory to include preload link headers
	// that don't have the corresponding allowed-alt-sxg with a valid
	// header-integrity.
	KeepNonSXGPreloads bool
}

// NewExchangeMetaFactory creates a new ExchangeMetaFactory.
func NewExchangeMetaFactory(c ExchangeConfig) *ExchangeMetaFactory {
	return &ExchangeMetaFactory{c}
}

// Get returns a new exchange.Factory set with the current AugmentedChain
// from e.CertManager.
func (e *ExchangeMetaFactory) Get() *exchange.Factory {
	chain := e.CertManager.GetAugmentedChain()

	// Use path.Join so the last path element in e.CertURLBase is kept
	// whether or not it has the trailing slash.
	urlPath := path.Join(e.CertURLBase.Path, chain.Digest)

	config := exchange.Config{
		Version:            e.Version,
		MIRecordSize:       e.MIRecordSize,
		CertChain:          chain,
		CertURL:            e.CertURLBase.ResolveReference(&url.URL{Path: urlPath}),
		PrivateKey:         e.PrivateKey,
		KeepNonSXGPreloads: e.KeepNonSXGPreloads,
	}
	return exchange.NewFactory(config)
}
