This repo is used to verify the download link for emeditor.com. Every 10 minutes, `main.go` is executed. It uses a browser to navigate to https://www.emeditor.com/download/, find the download hyperlink, then downloads the installer file. The script then uses `Get-AuthenticodeSignature` to validate the signature and check the expected fields of the signature.

The output of the script is written to [`status.json`](https://github.com/Emurasoft/signature-validation/blob/validation-results/status.json) in the `validation-results` branch.

If the signature is valid, `status.json` looks like this:

```json
{"result":{"valid":true},"time":"..."}
```

If the signature is invalid, then `status.json` will show `{"valid":false}` with a reason.

If the script had an error, it would look like:

```json
{"error":"..."}
```

If the signature was invalid or the script had an error, it creates a new issue to alert Makoto.