#! /bin/bash

go test \
  github.com/ervand7/go-musthave-diploma-tpl/internal/config \
  github.com/ervand7/go-musthave-diploma-tpl/internal/router \
  -count 1 -v -p 1

export DATABASE_URI='user=ervand password=ervand dbname=accrual_test host=localhost port=5432 sslmode=disable'
go test \
  github.com/ervand7/go-musthave-diploma-tpl/internal/database \
  github.com/ervand7/go-musthave-diploma-tpl/internal/views \
  github.com/ervand7/go-musthave-diploma-tpl/internal/utils \
  -count 1 -v -p 1