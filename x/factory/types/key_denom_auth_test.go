package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/x/factory/types"
)

func TestDenomAuthKey(t *testing.T) {
	tests := []struct {
		name  string
		denom string
		want  []byte
	}{
		{
			name:  "empty denom",
			denom: "",
			want:  []byte("/"),
		},
		{
			name:  "simple denom",
			denom: "abc",
			want:  []byte("abc/"),
		},
		{
			name:  "denom with special characters",
			denom: "coin/factory/creator/usdt",
			want:  []byte("coin/factory/creator/usdt/"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.DenomAuthKey(tt.denom)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAdminDenomAuthListKey(t *testing.T) {
	tests := []struct {
		name  string
		admin string
		want  []byte
	}{
		{
			name:  "empty admin",
			admin: "",
			want:  []byte("AdminDenomAuth/value//"),
		},
		{
			name:  "simple admin",
			admin: "admin1",
			want:  []byte("AdminDenomAuth/value/admin1/"),
		},
		{
			name:  "admin with special characters",
			admin: "zig1abc123",
			want:  []byte("AdminDenomAuth/value/zig1abc123/"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.AdminDenomAuthListKey(tt.admin)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDenomAuthNameKey(t *testing.T) {
	tests := []struct {
		name  string
		denom string
		want  []byte
	}{
		{
			name:  "simple denom",
			denom: "abc",
			want:  []byte("abc"),
		},
		{
			name:  "denom with special characters",
			denom: "coin/factory/creator/usdt",
			want:  []byte("coin/factory/creator/usdt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.DenomAuthNameKey(tt.denom)
			require.Equal(t, tt.want, got)
		})
	}
}
