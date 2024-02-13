### proton-drive

A library for interacting with the Proton Drive API.

This code was born out of (mild) frustration about the Proton Drive backend for rclone. Due to how the API works,
listing directory contents takes quite a lot of time. This, combined, with a lot of files makes sync operations
painfully slow.

To remedy this issue, this library takes a more unconventional approach: During initialization, the metadata of
**EVERY** file or directory is fetched and stored in a virtual file tree, that is kept up-to-date using Protons event
system.

While this means that directory listings are pretty much instant and don't require any calls to the API, it also means
that the startup time grows with the amount of files you have stored, and that the consuming app **MUST** be run as a
daemon.

#### Thanks

 * henrybear327 for publishing https://github.com/henrybear327/Proton-API-Bridge
 * Proton for publishing https://github.com/ProtonMail/go-proton-api
