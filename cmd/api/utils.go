package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/labstack/echo/v4"
)

func slowDown(secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
}


func PrintRoutes(e *echo.Echo) {
    fmt.Println("\n--- Registered Routes ---")
    w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
    
    // Header
    fmt.Fprintln(w, "METHOD\tPATH\tHANDLER")
    fmt.Fprintln(w, "------\t----\t-------")

    for _, route := range e.Routes() {
        if strings.Contains(route.Method, "echo_route_not_found") {
            continue
        }
        
        fmt.Fprintf(w, "%s\t%s\t%s\n", route.Method, route.Path, route.Name)
    }
    
    w.Flush()
    fmt.Println("-------------------------")
}

