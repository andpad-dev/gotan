[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Deliver an Encrypted File to the Audit Destination

When submitting customer data to an external auditor, we want the sender to know only the auditor's public key and ensure that only the auditor can open the contents. Let's use `crypto/hpke`, added in Go 1.26, to deliver a message following the standard library example. Let's investigate how to do this.

This is an exercise in investigating the roles of an API. In real-world use, always add an organizational security review for protocol selection and key management.

When you [run the following code in the Go Playground](https://go.dev/play/p/FDMWW7yXLFE), you can confirm that decryption succeeds with the same `info` and is rejected with a different `info`.

```go
package main

import (
	"crypto/hpke"
	"fmt"
)

func main() {
	kem, kdf, aead := hpke.MLKEM768X25519(), hpke.HKDFSHA256(), hpke.AES256GCM()
	privateKey, err := kem.GenerateKey()
	if err != nil {
		panic(err)
	}

	publicKey, err := kem.NewPublicKey(privateKey.PublicKey().Bytes())
	if err != nil {
		panic(err)
	}
	info := []byte("audit-export/v1")
	ciphertext, err := hpke.Seal(publicKey, kdf, aead, info, []byte("customer export ready"))
	if err != nil {
		panic(err)
	}

	plaintext, err := hpke.Open(privateKey, kdf, aead, info, ciphertext)
	if err != nil {
		panic(err)
	}
	_, wrongInfoErr := hpke.Open(privateKey, kdf, aead, []byte("audit-export/v2"), ciphertext)
	fmt.Printf("received=%q\n", plaintext)
	fmt.Println("different info rejected:", wrongInfoErr != nil)
}
```

Output:

```text
received="customer export ready"
different info rejected: true
```

---

## Question 1: What do the three components select?

What HPKE components are `kem`, `kdf`, and `aead`? Also give the specific implementation names selected in the sample.

<details>
<summary>Hint</summary>

- Use [How to investigate a category](../../README.md) as your starting point and read the Overview for `crypto/hpke`.
- Match the types `KEM`, `KDF`, and `AEAD` with the constructor names in the code.
- Separate the roles into “whose public key is used,” “how keys are derived from shared material,” and “how the body is encrypted and authenticated” to organize the similar abbreviations.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Confirm in the [Go 1.26 Release Notes](https://go.dev/doc/go1.26) that `crypto/hpke` is a new standard package.
2. Read the Overview and each type in [package crypto/hpke](https://pkg.go.dev/crypto/hpke).
3. Read [hpke.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=183) and the [standard library example](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke_test.go;l=21).

**Answer**

An HPKE cipher suite consists of a KEM (Key Encapsulation Mechanism), a KDF (Key Derivation Function), and an AEAD (Authenticated Encryption with Associated Data). The sample selects `MLKEM768X25519`, `HKDFSHA256`, and `AES256GCM`, respectively.

Rather than assembling cryptographic operations yourself, the starting point is to treat these components and the APIs defined by the package as one set.

</details>

---

## Question 2: What do the sender and receiver share?

Which key does the sender use with `Seal`, and which key does the receiver use with `Open`? Also, why must both pass the same `info`, and why does a different value cause failure?

<details>
<summary>Hint</summary>

- Place the arguments to `Seal` and `Open` side by side.
- Read the comments for the one-message `Seal` / `Open` API.
- Think of the public key as a mailbox that anyone can put something into, the private key as the key held only by the recipient, and `info` as a purpose label such as “audit export v1.” Confirm why the same ciphertext cannot be opened without using the same label.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Check the signatures of [Seal](https://pkg.go.dev/crypto/hpke#Seal) and [Open](https://pkg.go.dev/crypto/hpke#Open).
2. Read the comments for [`Seal` in hpke.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=183) and [`Open`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=224).

**Answer**

The sender seals with the recipient's public key, and the recipient opens with the corresponding private key. Here, the sender receives `privateKey.PublicKey().Bytes()` and reconstructs the public key with `kem.NewPublicKey`.

`info` is contextual information that both sides use with the same value. In the sample, it represents the version of the audit-data export format. Passing a different `info` to the decryption side causes `Open` to return an error because the ciphertext was not created in the same context.

</details>

---

## Question 3: Why use `kem.NewPublicKey`?

Why is a public-key byte sequence received over the wire reconstructed with `NewPublicKey` from the selected `kem`? Explain how to separate the KEM key format from the cipher suite consisting of KEM/KDF/AEAD, and how the combination should be handled.

<details>
<summary>Hint</summary>

- In the description of `KEM.NewPublicKey`, find how it interprets the byte sequence.
- Consider what happens if `MLKEM768X25519` is replaced with another KEM.
- Notice that neither the KDF nor the AEAD is passed to `NewPublicKey`, and confirm whether the cipher suite is automatically reconstructed from the public-key bytes.
- The same byte sequence is treated as a different format when read with another KEM. The KDF and AEAD, on the other hand, are selected separately. Try organizing this difference as a key format and a cipher-suite configuration.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read `NewPublicKey` on [KEM](https://pkg.go.dev/crypto/hpke#KEM).
2. In the [kem.go implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/kem.go;l=215), confirm that the KEM reconstructs the public key in its own format.
3. Return to the [`crypto/hpke` explanation for Go 1.26](https://go.dev/doc/go1.26).

**Answer**

A public-key byte sequence has no meaning by itself. `KEM.NewPublicKey` reconstructs and validates the sequence as a public key according to the selected KEM. This specifies the KEM key format; it does not recover or negotiate the KDF or AEAD selection.

The sender and receiver agree separately on the KEM used to encode the public key and on the cipher suite of KEM/KDF/AEAD as protocol information. They do not guess how to interpret the received byte sequence as a key for another scheme. Keeping scheme selection and key distribution from being mixed arbitrarily is a prerequisite for using the library API correctly.

</details>

---

<details>
<summary>Further note</summary>

For sending a single message, use `Seal` / `Open`. With `Sender` / `Recipient`, which send multiple messages, both sides must keep the order of successful `Seal` and `Open` calls in sync. Choose the API according to the use case.

</details>

---

## Investigation starting points

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [How to investigate 01-packages](../../README.md)
- [package crypto/hpke](https://pkg.go.dev/crypto/hpke)
