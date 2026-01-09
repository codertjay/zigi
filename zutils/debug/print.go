package d

import (
	"fmt"
	"reflect"
	"runtime"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fatih/color"
)

var step = 0

func P(args ...interface{}) {

	color.NoColor = false // Force color output

	_, file, line, ok := runtime.Caller(1) // Retrieves caller info
	if !ok {
		color.Red("Failed to obtain caller information")
		return
	}

	for _, arg := range args {
		typeOfArg := reflect.TypeOf(arg).String() // Gets the type of the argument
		//log.Printf("ZDEBUG: @ %s:%d\n", file, line)
		_, err := color.New(color.FgHiWhite).Add(color.Faint).Printf("ZDEBUG: @ %s:%d\n", file, line)
		if err != nil {
			return
		}
		//log.Printf("ZDEBUG: %s:\n", typeOfArg)
		_, err = color.New(color.FgGreen).Add(color.Faint).Printf("ZDEBUG: %s:\n", typeOfArg)
		if err != nil {
			return
		}
		//log.Printf("ZDEBUG: %v \n", arg)
		_, err = color.New(color.FgBlue, color.Bold).Printf("ZDEBUG: %v \n", arg)
		if err != nil {
			return
		}
	}
}

func L(ctx sdk.Context, msg string) string {

	step++
	return fmt.Sprintf(
		"[%d] %s \n context pointer: (%p), \n btime: %d \n bheight: %d",
		step,
		msg,
		&ctx,
		ctx.BlockTime().UnixNano(),
		ctx.BlockHeight(),
	)
}
