package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func validateToken(md metadata.MD, secret []byte) error {
	val := md["authorization"]
	if len(val) == 0 {
		return status.Error(codes.Unauthenticated, "missing auth header")
	}

	tokenStr := strings.TrimPrefix(val[0], "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return status.Error(codes.Unauthenticated, "token expired")
		}
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid token claims")
	}

	_, ok = claims["sub"].(float64)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid token claims")
	}

	if claims["type"] != "access" {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	return nil
}

func GRPCAuthInterceptor(secret []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		if err := validateToken(md, secret); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func GRPCStreamAuthInterceptor(secret []byte) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		if err := validateToken(md, secret); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}
