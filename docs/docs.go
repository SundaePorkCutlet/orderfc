// Package docs provides Swagger documentation for ORDERFC API.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "ORDERFC API",
        "description": "Order creation and history for Go Commerce.",
        "version": "1.0"
    },
    "host": "localhost:28082",
    "basePath": "/",
    "paths": {
        "/ping": {
            "get": {"summary": "Ping", "responses": {"200": {"description": "pong"}}}
        },
        "/health": {
            "get": {"summary": "Health", "responses": {"200": {"description": "healthy"}}}
        },
        "/api/v1/orders": {
            "post": {
                "security": [{"BearerAuth": []}],
                "summary": "Checkout (create order)",
                "parameters": [{
                    "in": "body",
                    "name": "body",
                    "required": true,
                    "schema": {
                        "type": "object",
                        "properties": {
                            "items": {"type": "array", "items": {"type": "object", "properties": {"product_id": {"type": "integer"}, "quantity": {"type": "integer"}, "price": {"type": "number"}}}},
                            "payment_method": {"type": "string"},
                            "shipping_address": {"type": "string"},
                            "idempotency_token": {"type": "string"}
                        }
                    }
                }],
                "responses": {"201": {"description": "order_id"}, "401": {"description": "Unauthorized"}}
            }
        },
        "/api/v1/orders/history": {
            "get": {
                "security": [{"BearerAuth": []}],
                "summary": "Get order history by user",
                "parameters": [{"in": "query", "name": "status", "type": "integer"}],
                "responses": {"200": {"description": "List of orders"}, "401": {"description": "Unauthorized"}}
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {"type": "apiKey", "name": "Authorization", "in": "header"}
    }
}`

func init() {
	swag.Register(swag.Name, &s{})
}

type s struct{}

func (s *s) ReadDoc() string {
	return docTemplate
}
