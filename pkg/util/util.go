package util

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// FatalIf exits if the error is not nil
func FatalIf(err error) {
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Fatal error: %s\n", err)
		os.Exit(-1)
	}
}

func FormatGrpcError(err error) gin.H{
	st, _ := status.FromError(err)
	return gin.H{
		"code": st.Code(),
		"message": st.Message(),
		"details": st.Details(),
	}
}
