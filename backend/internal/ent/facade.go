package ent

import "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"

type (
	Client      = generated.Client
	Tx          = generated.Tx
	Option      = generated.Option
	Mutation    = generated.Mutation
	Value       = generated.Value
	Hook        = generated.Hook
	Interceptor = generated.Interceptor
)

var (
	Driver       = generated.Driver
	Debug        = generated.Debug
	Log          = generated.Log
	ErrTxStarted = generated.ErrTxStarted
)

func NewClient(opts ...Option) *Client {
	return generated.NewClient(opts...)
}

func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	return generated.Open(driverName, dataSourceName, options...)
}
