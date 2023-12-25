package main

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/credentialprovider/v1"
)

type fakePlugin struct {
}

func (f *fakePlugin) GetCredentials(ctx context.Context, image string, args []string) (*v1.CredentialProviderResponse, error) {
	return &v1.CredentialProviderResponse{
		CacheKeyType:  v1.RegistryPluginCacheKeyType,
		CacheDuration: &metav1.Duration{Duration: 10 * time.Minute},
		Auth: map[string]v1.AuthConfig{
			"ghcr.io": {
				Username: "user",
				Password: "password",
			},
		},
	}, nil
}

func Test_runPlugin(t *testing.T) {
	testcases := []struct {
		name        string
		in          *bytes.Buffer
		expectedOut []byte
		expectErr   bool
	}{
		{
			name: "successful test case",
			in:   bytes.NewBufferString(`{"kind":"CredentialProviderRequest","apiVersion":"credentialprovider.kubelet.k8s.io/v1","image":"ghcr.io/foobar"}`),
			expectedOut: []byte(`{"kind":"CredentialProviderResponse","apiVersion":"credentialprovider.kubelet.k8s.io/v1","cacheKeyType":"Registry","cacheDuration":"10m0s","auth":{"ghcr.io":{"username":"user","password":"password"}}}
`),
			expectErr: false,
		},
		{
			name:        "invalid kind",
			in:          bytes.NewBufferString(`{"kind":"CredentialProviderFoo","apiVersion":"credentialprovider.kubelet.k8s.io/v1","image":"ghcr.io.io/foobar"}`),
			expectedOut: nil,
			expectErr:   true,
		},
		{
			name:        "invalid apiVersion",
			in:          bytes.NewBufferString(`{"kind":"CredentialProviderRequest","apiVersion":"foo.k8s.io/v1","image":"ghcr.io.io/foobar"}`),
			expectedOut: nil,
			expectErr:   true,
		},
		{
			name:        "empty image",
			in:          bytes.NewBufferString(`{"kind":"CredentialProviderRequest","apiVersion":"credentialprovider.kubelet.k8s.io/v1","image":""}`),
			expectedOut: nil,
			expectErr:   true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			p := NewCredentialProvider(&fakePlugin{})

			out := &bytes.Buffer{}
			err := p.runPlugin(context.TODO(), testcase.in, out, nil)
			if err != nil && !testcase.expectErr {
				t.Fatal(err)
			}

			if err == nil && testcase.expectErr {
				t.Error("expected error but got none")
			}

			if !reflect.DeepEqual(out.Bytes(), testcase.expectedOut) {
				t.Logf("actual output: %v", out.String())
				t.Logf("expected  output: %v", string(testcase.expectedOut))
				t.Errorf("unexpected output")
			}
		})
	}
}
