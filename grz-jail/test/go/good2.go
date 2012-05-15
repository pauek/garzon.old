package main

import ("fmt"; "os")

func main() {
   out, _ := os.Create("output")
   fmt.Fprintf(out, "Hello")
}
