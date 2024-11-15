package libkube

import (
	"reflect"
	"testing"

	discoveryv1 "k8s.io/api/discovery/v1"
)

func TestAddressType(t *testing.T) {
	t.Parallel()

	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    discoveryv1.AddressType
		wantErr bool
	}{
		{
			name:    "success case IPv4",
			args:    args{address: "10.1.1.1"},
			want:    discoveryv1.AddressTypeIPv4,
			wantErr: false,
		},
		{
			name:    "success case IPv6",
			args:    args{address: "fe80::857:d999:5316:78d4"},
			want:    discoveryv1.AddressTypeIPv6,
			wantErr: false,
		},
		{
			name:    "success case FQDN",
			args:    args{address: "a.domain.local"},
			want:    discoveryv1.AddressTypeFQDN,
			wantErr: false,
		},
		{
			name:    "error case",
			args:    args{address: "a.domain.local OR NOT!"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := AddressType(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressType() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressType() = %v, want %v", got, tt.want)
			}
		})
	}
}
