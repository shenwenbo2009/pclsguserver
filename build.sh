#!/bin/bash
set -e

echo "=== Building frontend ==="

echo "--- Building shared ---"
cd frontend/shared
npm install --silent

echo "--- Building admin ---"
cd ../admin
npm install --silent
npm run build

echo "--- Building market ---"
cd ../market
npm install --silent
npm run build

echo "=== Building Go binary ==="
cd ../..
go build -o pclsguserver ./cmd/server

echo "=== Done ==="
echo "Binary: ./pclsguserver"
echo "Run: ./pclsguserver -config config.yaml"