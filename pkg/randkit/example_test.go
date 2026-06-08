package randkit

import "fmt"

func ExampleStr() {
	fmt.Println(Str(WithSeed(1)))
	// Output:
	// qLKZasgepC
}

func ExampleStr_withLen() {
	fmt.Println(Str(WithSeed(1), WithLen(6)))
	// Output:
	// qLKZas
}

func ExampleStr_withChars() {
	fmt.Println(Str(WithChars(Digits), WithSeed(1), WithLen(8)))
	// Output:
	// 37790310
}

func ExampleStr_withPrefixSuffix() {
	fmt.Println(Str(WithPrefix("test-"), WithSuffix("-end"), WithLen(6), WithSeed(1)))
	// Output:
	// test-qLKZas-end
}

func ExampleFileName() {
	fmt.Println(FileName("/tmp", WithSeed(1)))
	// Output:
	// /tmp/file-qLKZasg.txt
}

func ExampleFileName_withExt() {
	fmt.Println(FileName("/tmp", WithExt(".json"), WithSeed(1)))
	// Output:
	// /tmp/file-qLKZasg.json
}

func ExampleInt() {
	fmt.Println(Int(100, WithSeed(1)))
	// Output:
	// 32
}

func ExamplePassword() {
	fmt.Println(Password(16, WithSeed(1)))
	// Output:
	// tSR9avhesITXkYun
}
