package types

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"helm.sh/helm/v3/pkg/repo"
	"time"
)

type ChartVersion struct {
	// Do not save ApplicationId into crd
	ApplicationId        string `json:"-"`
	ApplicationVersionId string `json:"verId"`
	repo.ChartVersion    `json:",inline"`
}

type Application struct {
	// application name
	Name          string `json:"name"`
	ApplicationId string `json:"appId"`
	// chart description
	Description string `json:"desc"`
	// application status
	Status string `json:"status"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`

	Charts []*ChartVersion `json:"charts"`
}

func (i *SavedIndex) Bytes() ([]byte, error) {

	d, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	w := zlib.NewWriter(buf)
	_, err = w.Write(d)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	encSrc := buf.Bytes()

	enc := base64.URLEncoding
	ret := make([]byte, enc.EncodedLen(len(encSrc)))

	enc.Encode(ret, encSrc)
	return ret, nil
}

type SavedIndex struct {
	APIVersion   string                  `json:"apiVersion"`
	Generated    time.Time               `json:"generated"`
	Applications map[string]*Application `json:"apps"`
	PublicKeys   []string                `json:"publicKeys,omitempty"`

	// Annotations are additional mappings uninterpreted by Helm. They are made available for
	// other applications to add information to the index file.
	Annotations map[string]string `json:"annotations,omitempty"`
}
