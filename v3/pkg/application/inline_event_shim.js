// Inline Wails events shim for InitialHTML windows.
//
// Pages loaded via WebviewWindowOptions.HTML are served with
// `window.location.origin === "null"`, so the modern HTTP runtime at
// /wails/runtime.js can never be imported (fetch fails). This shim
// installs the minimum subset of the runtime that postMessage-based
// custom-event traffic needs:
//
//   * window._wails.dispatchWailsEvent  — the framework calls this when
//     the host emits a custom event to the page.
//   * window.wails.Events.On(name, cb)  — the page subscribes.
//   * window.wails.Events.Emit(name)    — the page fires a bare-name
//     event back to the host. Routed through
//     `window._wails.invoke("wails:event:emit:" + name)` which the
//     framework forwards if the owning window has
//     WebviewWindowOptions.AllowSimpleEventEmit set.
//
// Once the platform layer has injected window._wails.invoke the shim
// fires `wails:runtime:ready` so any pending host-side queued events
// flush, and then reports each event name it has a listener for with
// `wails:event:listen:<name>` (and `wails:event:unlisten:<name>` when the
// last listener of a name goes), because the host evaluates an event only
// in a page that reported a listener for its name. If the modern HTTP runtime later loads it will overwrite
// window.wails.Events with its richer implementation — that's fine,
// our subset is the floor not the ceiling.
//
// Kept in ES5-compatible syntax (no const/let/arrow) so older WebView
// engines that may still surface in unusual platform configurations
// don't fail on parse.
(function () {
    var w = window._wails = window._wails || {};
    if (window.wails && window.wails.Events) {
        return; // a full runtime is already in scope
    }
    var listeners = Object.create(null);
    var ready = false;
    function announce(kind, name) {
        if (ready) w.invoke("wails:event:" + kind + ":" + name);
    }
    w.dispatchWailsEvent = w.dispatchWailsEvent || function (ev) {
        if (!ev || !ev.name) return;
        var cbs = listeners[ev.name];
        if (!cbs) return;
        for (var i = 0; i < cbs.length; i++) {
            try { cbs[i](ev); } catch (_) { /* swallow handler errors */ }
        }
    };
    window.wails = window.wails || {};
    window.wails.Events = {
        On: function (name, cb) {
            if (!listeners[name]) {
                listeners[name] = [];
                announce("listen", name);
            }
            listeners[name].push(cb);
            return function () {
                var arr = listeners[name];
                if (!arr) return;
                var i = arr.indexOf(cb);
                if (i >= 0) arr.splice(i, 1);
                if (arr.length === 0) {
                    delete listeners[name];
                    announce("unlisten", name);
                }
            };
        },
        Emit: function (eventOrName) {
            var name = (typeof eventOrName === "string")
                ? eventOrName
                : (eventOrName && eventOrName.name);
            if (!name || typeof w.invoke !== "function") return;
            w.invoke("wails:event:emit:" + name);
        },
    };
    (function waitForInvoke() {
        if (typeof w.invoke === "function") {
            w.invoke("wails:runtime:ready");
            ready = true;
            for (var name in listeners) announce("listen", name);
        } else {
            setTimeout(waitForInvoke, 30);
        }
    })();
})();
