# Unreleased Changes

<!-- 
This file is used to collect changelog entries for the next v3 release.
Add your changes under the appropriate sections below.

Guidelines:
- Follow the "Keep a Changelog" format (https://keepachangelog.com/)
- Write clear, concise descriptions of changes
- Include the impact on users when relevant
- Use present tense ("Add feature" not "Added feature")
- Reference issue/PR numbers when applicable

This file is automatically processed by the nightly release workflow.
After processing, the content will be moved to the main changelog and this file will be reset.
-->

## Added
<!-- New features, capabilities, or enhancements -->
- Add `MacOptions.DisableActivationOnLaunch`, which stops a regular application from activating itself when it finishes launching, so an application launched in the background (`open -g`) stays behind the frontmost application

## Changed
<!-- Changes in existing functionality -->
- Evaluate an event in a window only when the runtime of its page has a listener for the event name: the runtime reports the first listener of a name with `wails:event:listen:<name>` and the removal of the last one with `wails:event:unlisten:<name>`, and an event emitted before the runtime is ready is kept and flushed as before. A page without a listener for an event no longer evaluates it, which on macOS left about 1.4 KB of WebKit garbage per event, and native window events reach only pages that listen to them

## Fixed
<!-- Bug fixes -->

## Deprecated
<!-- Soon-to-be removed features -->

## Removed
<!-- Features removed in this release -->

## Security
<!-- Security-related changes -->

---

### Example Entries:

**Added:**
- Add support for custom window icons in application options
- Add new `SetWindowIcon()` method to runtime API (#1234)

**Changed:**
- Update minimum Go version requirement to 1.21
- Improve error messages for invalid configuration files

**Fixed:**
- Fix memory leak in event system during window close operations (#5678)
- Fix crash when using context menus on Linux with Wayland

**Security:**
- Update dependencies to address CVE-2024-12345 in third-party library
