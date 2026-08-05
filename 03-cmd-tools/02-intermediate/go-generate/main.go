package main

//go:generate echo "hello from go generate"
//go:generate date "+%Y-%m-%d"
//go:generate sh -c "echo GOFILE=$GOFILE GOLINE=$GOLINE GOPACKAGE=$GOPACKAGE"
//go:generate -command say echo
//go:generate say "-command による別名も使える"

func main() {}
