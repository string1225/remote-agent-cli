var zf = { exports: {} }, ku = {};
var Td;
function I0() {
  if (Td) return ku;
  Td = 1;
  var r = /* @__PURE__ */ Symbol.for("react.transitional.element"), x = /* @__PURE__ */ Symbol.for("react.fragment");
  function A(o, _, D) {
    var U = null;
    if (D !== void 0 && (U = "" + D), _.key !== void 0 && (U = "" + _.key), "key" in _) {
      D = {};
      for (var O in _)
        O !== "key" && (D[O] = _[O]);
    } else D = _;
    return _ = D.ref, {
      $$typeof: r,
      type: o,
      key: U,
      ref: _ !== void 0 ? _ : null,
      props: D
    };
  }
  return ku.Fragment = x, ku.jsx = A, ku.jsxs = A, ku;
}
var Ad;
function t1() {
  return Ad || (Ad = 1, zf.exports = I0()), zf.exports;
}
var c = t1(), Mf = { exports: {} }, $u = {}, Nf = { exports: {} }, Df = {};
var Ed;
function l1() {
  return Ed || (Ed = 1, (function(r) {
    function x(p, H) {
      var w = p.length;
      p.push(H);
      t: for (; 0 < w; ) {
        var ot = w - 1 >>> 1, d = p[ot];
        if (0 < _(d, H))
          p[ot] = H, p[w] = d, w = ot;
        else break t;
      }
    }
    function A(p) {
      return p.length === 0 ? null : p[0];
    }
    function o(p) {
      if (p.length === 0) return null;
      var H = p[0], w = p.pop();
      if (w !== H) {
        p[0] = w;
        t: for (var ot = 0, d = p.length, M = d >>> 1; ot < M; ) {
          var Y = 2 * (ot + 1) - 1, q = p[Y], L = Y + 1, it = p[L];
          if (0 > _(q, w))
            L < d && 0 > _(it, q) ? (p[ot] = it, p[L] = w, ot = L) : (p[ot] = q, p[Y] = w, ot = Y);
          else if (L < d && 0 > _(it, w))
            p[ot] = it, p[L] = w, ot = L;
          else break t;
        }
      }
      return H;
    }
    function _(p, H) {
      var w = p.sortIndex - H.sortIndex;
      return w !== 0 ? w : p.id - H.id;
    }
    if (r.unstable_now = void 0, typeof performance == "object" && typeof performance.now == "function") {
      var D = performance;
      r.unstable_now = function() {
        return D.now();
      };
    } else {
      var U = Date, O = U.now();
      r.unstable_now = function() {
        return U.now() - O;
      };
    }
    var N = [], b = [], C = 1, G = null, W = 3, st = !1, B = !1, rt = !1, Ht = !1, xt = typeof setTimeout == "function" ? setTimeout : null, qt = typeof clearTimeout == "function" ? clearTimeout : null, Et = typeof setImmediate < "u" ? setImmediate : null;
    function nt(p) {
      for (var H = A(b); H !== null; ) {
        if (H.callback === null) o(b);
        else if (H.startTime <= p)
          o(b), H.sortIndex = H.expirationTime, x(N, H);
        else break;
        H = A(b);
      }
    }
    function P(p) {
      if (rt = !1, nt(p), !B)
        if (A(N) !== null)
          B = !0, $t || ($t = !0, Zt());
        else {
          var H = A(b);
          H !== null && Dt(P, H.startTime - p);
        }
    }
    var $t = !1, V = -1, zt = 5, Wt = -1;
    function Fl() {
      return Ht ? !0 : !(r.unstable_now() - Wt < zt);
    }
    function nl() {
      if (Ht = !1, $t) {
        var p = r.unstable_now();
        Wt = p;
        var H = !0;
        try {
          t: {
            B = !1, rt && (rt = !1, qt(V), V = -1), st = !0;
            var w = W;
            try {
              l: {
                for (nt(p), G = A(N); G !== null && !(G.expirationTime > p && Fl()); ) {
                  var ot = G.callback;
                  if (typeof ot == "function") {
                    G.callback = null, W = G.priorityLevel;
                    var d = ot(
                      G.expirationTime <= p
                    );
                    if (p = r.unstable_now(), typeof d == "function") {
                      G.callback = d, nt(p), H = !0;
                      break l;
                    }
                    G === A(N) && o(N), nt(p);
                  } else o(N);
                  G = A(N);
                }
                if (G !== null) H = !0;
                else {
                  var M = A(b);
                  M !== null && Dt(
                    P,
                    M.startTime - p
                  ), H = !1;
                }
              }
              break t;
            } finally {
              G = null, W = w, st = !1;
            }
            H = void 0;
          }
        } finally {
          H ? Zt() : $t = !1;
        }
      }
    }
    var Zt;
    if (typeof Et == "function")
      Zt = function() {
        Et(nl);
      };
    else if (typeof MessageChannel < "u") {
      var Ql = new MessageChannel(), vl = Ql.port2;
      Ql.port1.onmessage = nl, Zt = function() {
        vl.postMessage(null);
      };
    } else
      Zt = function() {
        xt(nl, 0);
      };
    function Dt(p, H) {
      V = xt(function() {
        p(r.unstable_now());
      }, H);
    }
    r.unstable_IdlePriority = 5, r.unstable_ImmediatePriority = 1, r.unstable_LowPriority = 4, r.unstable_NormalPriority = 3, r.unstable_Profiling = null, r.unstable_UserBlockingPriority = 2, r.unstable_cancelCallback = function(p) {
      p.callback = null;
    }, r.unstable_forceFrameRate = function(p) {
      0 > p || 125 < p ? console.error(
        "forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"
      ) : zt = 0 < p ? Math.floor(1e3 / p) : 5;
    }, r.unstable_getCurrentPriorityLevel = function() {
      return W;
    }, r.unstable_next = function(p) {
      switch (W) {
        case 1:
        case 2:
        case 3:
          var H = 3;
          break;
        default:
          H = W;
      }
      var w = W;
      W = H;
      try {
        return p();
      } finally {
        W = w;
      }
    }, r.unstable_requestPaint = function() {
      Ht = !0;
    }, r.unstable_runWithPriority = function(p, H) {
      switch (p) {
        case 1:
        case 2:
        case 3:
        case 4:
        case 5:
          break;
        default:
          p = 3;
      }
      var w = W;
      W = p;
      try {
        return H();
      } finally {
        W = w;
      }
    }, r.unstable_scheduleCallback = function(p, H, w) {
      var ot = r.unstable_now();
      switch (typeof w == "object" && w !== null ? (w = w.delay, w = typeof w == "number" && 0 < w ? ot + w : ot) : w = ot, p) {
        case 1:
          var d = -1;
          break;
        case 2:
          d = 250;
          break;
        case 5:
          d = 1073741823;
          break;
        case 4:
          d = 1e4;
          break;
        default:
          d = 5e3;
      }
      return d = w + d, p = {
        id: C++,
        callback: H,
        priorityLevel: p,
        startTime: w,
        expirationTime: d,
        sortIndex: -1
      }, w > ot ? (p.sortIndex = w, x(b, p), A(N) === null && p === A(b) && (rt ? (qt(V), V = -1) : rt = !0, Dt(P, w - ot))) : (p.sortIndex = d, x(N, p), B || st || (B = !0, $t || ($t = !0, Zt()))), p;
    }, r.unstable_shouldYield = Fl, r.unstable_wrapCallback = function(p) {
      var H = W;
      return function() {
        var w = W;
        W = H;
        try {
          return p.apply(this, arguments);
        } finally {
          W = w;
        }
      };
    };
  })(Df)), Df;
}
var zd;
function e1() {
  return zd || (zd = 1, Nf.exports = l1()), Nf.exports;
}
var Of = { exports: {} }, I = {};
var Md;
function a1() {
  if (Md) return I;
  Md = 1;
  var r = /* @__PURE__ */ Symbol.for("react.transitional.element"), x = /* @__PURE__ */ Symbol.for("react.portal"), A = /* @__PURE__ */ Symbol.for("react.fragment"), o = /* @__PURE__ */ Symbol.for("react.strict_mode"), _ = /* @__PURE__ */ Symbol.for("react.profiler"), D = /* @__PURE__ */ Symbol.for("react.consumer"), U = /* @__PURE__ */ Symbol.for("react.context"), O = /* @__PURE__ */ Symbol.for("react.forward_ref"), N = /* @__PURE__ */ Symbol.for("react.suspense"), b = /* @__PURE__ */ Symbol.for("react.memo"), C = /* @__PURE__ */ Symbol.for("react.lazy"), G = Symbol.iterator;
  function W(d) {
    return d === null || typeof d != "object" ? null : (d = G && d[G] || d["@@iterator"], typeof d == "function" ? d : null);
  }
  var st = {
    isMounted: function() {
      return !1;
    },
    enqueueForceUpdate: function() {
    },
    enqueueReplaceState: function() {
    },
    enqueueSetState: function() {
    }
  }, B = Object.assign, rt = {};
  function Ht(d, M, Y) {
    this.props = d, this.context = M, this.refs = rt, this.updater = Y || st;
  }
  Ht.prototype.isReactComponent = {}, Ht.prototype.setState = function(d, M) {
    if (typeof d != "object" && typeof d != "function" && d != null)
      throw Error(
        "takes an object of state variables to update or a function which returns an object of state variables."
      );
    this.updater.enqueueSetState(this, d, M, "setState");
  }, Ht.prototype.forceUpdate = function(d) {
    this.updater.enqueueForceUpdate(this, d, "forceUpdate");
  };
  function xt() {
  }
  xt.prototype = Ht.prototype;
  function qt(d, M, Y) {
    this.props = d, this.context = M, this.refs = rt, this.updater = Y || st;
  }
  var Et = qt.prototype = new xt();
  Et.constructor = qt, B(Et, Ht.prototype), Et.isPureReactComponent = !0;
  var nt = Array.isArray, P = { H: null, A: null, T: null, S: null, V: null }, $t = Object.prototype.hasOwnProperty;
  function V(d, M, Y, q, L, it) {
    return Y = it.ref, {
      $$typeof: r,
      type: d,
      key: M,
      ref: Y !== void 0 ? Y : null,
      props: it
    };
  }
  function zt(d, M) {
    return V(
      d.type,
      M,
      void 0,
      void 0,
      void 0,
      d.props
    );
  }
  function Wt(d) {
    return typeof d == "object" && d !== null && d.$$typeof === r;
  }
  function Fl(d) {
    var M = { "=": "=0", ":": "=2" };
    return "$" + d.replace(/[=:]/g, function(Y) {
      return M[Y];
    });
  }
  var nl = /\/+/g;
  function Zt(d, M) {
    return typeof d == "object" && d !== null && d.key != null ? Fl("" + d.key) : M.toString(36);
  }
  function Ql() {
  }
  function vl(d) {
    switch (d.status) {
      case "fulfilled":
        return d.value;
      case "rejected":
        throw d.reason;
      default:
        switch (typeof d.status == "string" ? d.then(Ql, Ql) : (d.status = "pending", d.then(
          function(M) {
            d.status === "pending" && (d.status = "fulfilled", d.value = M);
          },
          function(M) {
            d.status === "pending" && (d.status = "rejected", d.reason = M);
          }
        )), d.status) {
          case "fulfilled":
            return d.value;
          case "rejected":
            throw d.reason;
        }
    }
    throw d;
  }
  function Dt(d, M, Y, q, L) {
    var it = typeof d;
    (it === "undefined" || it === "boolean") && (d = null);
    var $ = !1;
    if (d === null) $ = !0;
    else
      switch (it) {
        case "bigint":
        case "string":
        case "number":
          $ = !0;
          break;
        case "object":
          switch (d.$$typeof) {
            case r:
            case x:
              $ = !0;
              break;
            case C:
              return $ = d._init, Dt(
                $(d._payload),
                M,
                Y,
                q,
                L
              );
          }
      }
    if ($)
      return L = L(d), $ = q === "" ? "." + Zt(d, 0) : q, nt(L) ? (Y = "", $ != null && (Y = $.replace(nl, "$&/") + "/"), Dt(L, M, Y, "", function(Zl) {
        return Zl;
      })) : L != null && (Wt(L) && (L = zt(
        L,
        Y + (L.key == null || d && d.key === L.key ? "" : ("" + L.key).replace(
          nl,
          "$&/"
        ) + "/") + $
      )), M.push(L)), 1;
    $ = 0;
    var ll = q === "" ? "." : q + ":";
    if (nt(d))
      for (var Tt = 0; Tt < d.length; Tt++)
        q = d[Tt], it = ll + Zt(q, Tt), $ += Dt(
          q,
          M,
          Y,
          it,
          L
        );
    else if (Tt = W(d), typeof Tt == "function")
      for (d = Tt.call(d), Tt = 0; !(q = d.next()).done; )
        q = q.value, it = ll + Zt(q, Tt++), $ += Dt(
          q,
          M,
          Y,
          it,
          L
        );
    else if (it === "object") {
      if (typeof d.then == "function")
        return Dt(
          vl(d),
          M,
          Y,
          q,
          L
        );
      throw M = String(d), Error(
        "Objects are not valid as a React child (found: " + (M === "[object Object]" ? "object with keys {" + Object.keys(d).join(", ") + "}" : M) + "). If you meant to render a collection of children, use an array instead."
      );
    }
    return $;
  }
  function p(d, M, Y) {
    if (d == null) return d;
    var q = [], L = 0;
    return Dt(d, q, "", "", function(it) {
      return M.call(Y, it, L++);
    }), q;
  }
  function H(d) {
    if (d._status === -1) {
      var M = d._result;
      M = M(), M.then(
        function(Y) {
          (d._status === 0 || d._status === -1) && (d._status = 1, d._result = Y);
        },
        function(Y) {
          (d._status === 0 || d._status === -1) && (d._status = 2, d._result = Y);
        }
      ), d._status === -1 && (d._status = 0, d._result = M);
    }
    if (d._status === 1) return d._result.default;
    throw d._result;
  }
  var w = typeof reportError == "function" ? reportError : function(d) {
    if (typeof window == "object" && typeof window.ErrorEvent == "function") {
      var M = new window.ErrorEvent("error", {
        bubbles: !0,
        cancelable: !0,
        message: typeof d == "object" && d !== null && typeof d.message == "string" ? String(d.message) : String(d),
        error: d
      });
      if (!window.dispatchEvent(M)) return;
    } else if (typeof process == "object" && typeof process.emit == "function") {
      process.emit("uncaughtException", d);
      return;
    }
    console.error(d);
  };
  function ot() {
  }
  return I.Children = {
    map: p,
    forEach: function(d, M, Y) {
      p(
        d,
        function() {
          M.apply(this, arguments);
        },
        Y
      );
    },
    count: function(d) {
      var M = 0;
      return p(d, function() {
        M++;
      }), M;
    },
    toArray: function(d) {
      return p(d, function(M) {
        return M;
      }) || [];
    },
    only: function(d) {
      if (!Wt(d))
        throw Error(
          "React.Children.only expected to receive a single React element child."
        );
      return d;
    }
  }, I.Component = Ht, I.Fragment = A, I.Profiler = _, I.PureComponent = qt, I.StrictMode = o, I.Suspense = N, I.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = P, I.__COMPILER_RUNTIME = {
    __proto__: null,
    c: function(d) {
      return P.H.useMemoCache(d);
    }
  }, I.cache = function(d) {
    return function() {
      return d.apply(null, arguments);
    };
  }, I.cloneElement = function(d, M, Y) {
    if (d == null)
      throw Error(
        "The argument must be a React element, but you passed " + d + "."
      );
    var q = B({}, d.props), L = d.key, it = void 0;
    if (M != null)
      for ($ in M.ref !== void 0 && (it = void 0), M.key !== void 0 && (L = "" + M.key), M)
        !$t.call(M, $) || $ === "key" || $ === "__self" || $ === "__source" || $ === "ref" && M.ref === void 0 || (q[$] = M[$]);
    var $ = arguments.length - 2;
    if ($ === 1) q.children = Y;
    else if (1 < $) {
      for (var ll = Array($), Tt = 0; Tt < $; Tt++)
        ll[Tt] = arguments[Tt + 2];
      q.children = ll;
    }
    return V(d.type, L, void 0, void 0, it, q);
  }, I.createContext = function(d) {
    return d = {
      $$typeof: U,
      _currentValue: d,
      _currentValue2: d,
      _threadCount: 0,
      Provider: null,
      Consumer: null
    }, d.Provider = d, d.Consumer = {
      $$typeof: D,
      _context: d
    }, d;
  }, I.createElement = function(d, M, Y) {
    var q, L = {}, it = null;
    if (M != null)
      for (q in M.key !== void 0 && (it = "" + M.key), M)
        $t.call(M, q) && q !== "key" && q !== "__self" && q !== "__source" && (L[q] = M[q]);
    var $ = arguments.length - 2;
    if ($ === 1) L.children = Y;
    else if (1 < $) {
      for (var ll = Array($), Tt = 0; Tt < $; Tt++)
        ll[Tt] = arguments[Tt + 2];
      L.children = ll;
    }
    if (d && d.defaultProps)
      for (q in $ = d.defaultProps, $)
        L[q] === void 0 && (L[q] = $[q]);
    return V(d, it, void 0, void 0, null, L);
  }, I.createRef = function() {
    return { current: null };
  }, I.forwardRef = function(d) {
    return { $$typeof: O, render: d };
  }, I.isValidElement = Wt, I.lazy = function(d) {
    return {
      $$typeof: C,
      _payload: { _status: -1, _result: d },
      _init: H
    };
  }, I.memo = function(d, M) {
    return {
      $$typeof: b,
      type: d,
      compare: M === void 0 ? null : M
    };
  }, I.startTransition = function(d) {
    var M = P.T, Y = {};
    P.T = Y;
    try {
      var q = d(), L = P.S;
      L !== null && L(Y, q), typeof q == "object" && q !== null && typeof q.then == "function" && q.then(ot, w);
    } catch (it) {
      w(it);
    } finally {
      P.T = M;
    }
  }, I.unstable_useCacheRefresh = function() {
    return P.H.useCacheRefresh();
  }, I.use = function(d) {
    return P.H.use(d);
  }, I.useActionState = function(d, M, Y) {
    return P.H.useActionState(d, M, Y);
  }, I.useCallback = function(d, M) {
    return P.H.useCallback(d, M);
  }, I.useContext = function(d) {
    return P.H.useContext(d);
  }, I.useDebugValue = function() {
  }, I.useDeferredValue = function(d, M) {
    return P.H.useDeferredValue(d, M);
  }, I.useEffect = function(d, M, Y) {
    var q = P.H;
    if (typeof Y == "function")
      throw Error(
        "useEffect CRUD overload is not enabled in this build of React."
      );
    return q.useEffect(d, M);
  }, I.useId = function() {
    return P.H.useId();
  }, I.useImperativeHandle = function(d, M, Y) {
    return P.H.useImperativeHandle(d, M, Y);
  }, I.useInsertionEffect = function(d, M) {
    return P.H.useInsertionEffect(d, M);
  }, I.useLayoutEffect = function(d, M) {
    return P.H.useLayoutEffect(d, M);
  }, I.useMemo = function(d, M) {
    return P.H.useMemo(d, M);
  }, I.useOptimistic = function(d, M) {
    return P.H.useOptimistic(d, M);
  }, I.useReducer = function(d, M, Y) {
    return P.H.useReducer(d, M, Y);
  }, I.useRef = function(d) {
    return P.H.useRef(d);
  }, I.useState = function(d) {
    return P.H.useState(d);
  }, I.useSyncExternalStore = function(d, M, Y) {
    return P.H.useSyncExternalStore(
      d,
      M,
      Y
    );
  }, I.useTransition = function() {
    return P.H.useTransition();
  }, I.version = "19.1.1", I;
}
var Nd;
function Uf() {
  return Nd || (Nd = 1, Of.exports = a1()), Of.exports;
}
var Rf = { exports: {} }, ul = {};
var Dd;
function u1() {
  if (Dd) return ul;
  Dd = 1;
  var r = Uf();
  function x(N) {
    var b = "https://react.dev/errors/" + N;
    if (1 < arguments.length) {
      b += "?args[]=" + encodeURIComponent(arguments[1]);
      for (var C = 2; C < arguments.length; C++)
        b += "&args[]=" + encodeURIComponent(arguments[C]);
    }
    return "Minified React error #" + N + "; visit " + b + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
  }
  function A() {
  }
  var o = {
    d: {
      f: A,
      r: function() {
        throw Error(x(522));
      },
      D: A,
      C: A,
      L: A,
      m: A,
      X: A,
      S: A,
      M: A
    },
    p: 0,
    findDOMNode: null
  }, _ = /* @__PURE__ */ Symbol.for("react.portal");
  function D(N, b, C) {
    var G = 3 < arguments.length && arguments[3] !== void 0 ? arguments[3] : null;
    return {
      $$typeof: _,
      key: G == null ? null : "" + G,
      children: N,
      containerInfo: b,
      implementation: C
    };
  }
  var U = r.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;
  function O(N, b) {
    if (N === "font") return "";
    if (typeof b == "string")
      return b === "use-credentials" ? b : "";
  }
  return ul.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = o, ul.createPortal = function(N, b) {
    var C = 2 < arguments.length && arguments[2] !== void 0 ? arguments[2] : null;
    if (!b || b.nodeType !== 1 && b.nodeType !== 9 && b.nodeType !== 11)
      throw Error(x(299));
    return D(N, b, null, C);
  }, ul.flushSync = function(N) {
    var b = U.T, C = o.p;
    try {
      if (U.T = null, o.p = 2, N) return N();
    } finally {
      U.T = b, o.p = C, o.d.f();
    }
  }, ul.preconnect = function(N, b) {
    typeof N == "string" && (b ? (b = b.crossOrigin, b = typeof b == "string" ? b === "use-credentials" ? b : "" : void 0) : b = null, o.d.C(N, b));
  }, ul.prefetchDNS = function(N) {
    typeof N == "string" && o.d.D(N);
  }, ul.preinit = function(N, b) {
    if (typeof N == "string" && b && typeof b.as == "string") {
      var C = b.as, G = O(C, b.crossOrigin), W = typeof b.integrity == "string" ? b.integrity : void 0, st = typeof b.fetchPriority == "string" ? b.fetchPriority : void 0;
      C === "style" ? o.d.S(
        N,
        typeof b.precedence == "string" ? b.precedence : void 0,
        {
          crossOrigin: G,
          integrity: W,
          fetchPriority: st
        }
      ) : C === "script" && o.d.X(N, {
        crossOrigin: G,
        integrity: W,
        fetchPriority: st,
        nonce: typeof b.nonce == "string" ? b.nonce : void 0
      });
    }
  }, ul.preinitModule = function(N, b) {
    if (typeof N == "string")
      if (typeof b == "object" && b !== null) {
        if (b.as == null || b.as === "script") {
          var C = O(
            b.as,
            b.crossOrigin
          );
          o.d.M(N, {
            crossOrigin: C,
            integrity: typeof b.integrity == "string" ? b.integrity : void 0,
            nonce: typeof b.nonce == "string" ? b.nonce : void 0
          });
        }
      } else b == null && o.d.M(N);
  }, ul.preload = function(N, b) {
    if (typeof N == "string" && typeof b == "object" && b !== null && typeof b.as == "string") {
      var C = b.as, G = O(C, b.crossOrigin);
      o.d.L(N, C, {
        crossOrigin: G,
        integrity: typeof b.integrity == "string" ? b.integrity : void 0,
        nonce: typeof b.nonce == "string" ? b.nonce : void 0,
        type: typeof b.type == "string" ? b.type : void 0,
        fetchPriority: typeof b.fetchPriority == "string" ? b.fetchPriority : void 0,
        referrerPolicy: typeof b.referrerPolicy == "string" ? b.referrerPolicy : void 0,
        imageSrcSet: typeof b.imageSrcSet == "string" ? b.imageSrcSet : void 0,
        imageSizes: typeof b.imageSizes == "string" ? b.imageSizes : void 0,
        media: typeof b.media == "string" ? b.media : void 0
      });
    }
  }, ul.preloadModule = function(N, b) {
    if (typeof N == "string")
      if (b) {
        var C = O(b.as, b.crossOrigin);
        o.d.m(N, {
          as: typeof b.as == "string" && b.as !== "script" ? b.as : void 0,
          crossOrigin: C,
          integrity: typeof b.integrity == "string" ? b.integrity : void 0
        });
      } else o.d.m(N);
  }, ul.requestFormReset = function(N) {
    o.d.r(N);
  }, ul.unstable_batchedUpdates = function(N, b) {
    return N(b);
  }, ul.useFormState = function(N, b, C) {
    return U.H.useFormState(N, b, C);
  }, ul.useFormStatus = function() {
    return U.H.useHostTransitionStatus();
  }, ul.version = "19.1.1", ul;
}
var Od;
function n1() {
  if (Od) return Rf.exports;
  Od = 1;
  function r() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(r);
      } catch (x) {
        console.error(x);
      }
  }
  return r(), Rf.exports = u1(), Rf.exports;
}
var Rd;
function i1() {
  if (Rd) return $u;
  Rd = 1;
  var r = e1(), x = Uf(), A = n1();
  function o(t) {
    var l = "https://react.dev/errors/" + t;
    if (1 < arguments.length) {
      l += "?args[]=" + encodeURIComponent(arguments[1]);
      for (var e = 2; e < arguments.length; e++)
        l += "&args[]=" + encodeURIComponent(arguments[e]);
    }
    return "Minified React error #" + t + "; visit " + l + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
  }
  function _(t) {
    return !(!t || t.nodeType !== 1 && t.nodeType !== 9 && t.nodeType !== 11);
  }
  function D(t) {
    var l = t, e = t;
    if (t.alternate) for (; l.return; ) l = l.return;
    else {
      t = l;
      do
        l = t, (l.flags & 4098) !== 0 && (e = l.return), t = l.return;
      while (t);
    }
    return l.tag === 3 ? e : null;
  }
  function U(t) {
    if (t.tag === 13) {
      var l = t.memoizedState;
      if (l === null && (t = t.alternate, t !== null && (l = t.memoizedState)), l !== null) return l.dehydrated;
    }
    return null;
  }
  function O(t) {
    if (D(t) !== t)
      throw Error(o(188));
  }
  function N(t) {
    var l = t.alternate;
    if (!l) {
      if (l = D(t), l === null) throw Error(o(188));
      return l !== t ? null : t;
    }
    for (var e = t, a = l; ; ) {
      var u = e.return;
      if (u === null) break;
      var n = u.alternate;
      if (n === null) {
        if (a = u.return, a !== null) {
          e = a;
          continue;
        }
        break;
      }
      if (u.child === n.child) {
        for (n = u.child; n; ) {
          if (n === e) return O(u), t;
          if (n === a) return O(u), l;
          n = n.sibling;
        }
        throw Error(o(188));
      }
      if (e.return !== a.return) e = u, a = n;
      else {
        for (var i = !1, f = u.child; f; ) {
          if (f === e) {
            i = !0, e = u, a = n;
            break;
          }
          if (f === a) {
            i = !0, a = u, e = n;
            break;
          }
          f = f.sibling;
        }
        if (!i) {
          for (f = n.child; f; ) {
            if (f === e) {
              i = !0, e = n, a = u;
              break;
            }
            if (f === a) {
              i = !0, a = n, e = u;
              break;
            }
            f = f.sibling;
          }
          if (!i) throw Error(o(189));
        }
      }
      if (e.alternate !== a) throw Error(o(190));
    }
    if (e.tag !== 3) throw Error(o(188));
    return e.stateNode.current === e ? t : l;
  }
  function b(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t;
    for (t = t.child; t !== null; ) {
      if (l = b(t), l !== null) return l;
      t = t.sibling;
    }
    return null;
  }
  var C = Object.assign, G = /* @__PURE__ */ Symbol.for("react.element"), W = /* @__PURE__ */ Symbol.for("react.transitional.element"), st = /* @__PURE__ */ Symbol.for("react.portal"), B = /* @__PURE__ */ Symbol.for("react.fragment"), rt = /* @__PURE__ */ Symbol.for("react.strict_mode"), Ht = /* @__PURE__ */ Symbol.for("react.profiler"), xt = /* @__PURE__ */ Symbol.for("react.provider"), qt = /* @__PURE__ */ Symbol.for("react.consumer"), Et = /* @__PURE__ */ Symbol.for("react.context"), nt = /* @__PURE__ */ Symbol.for("react.forward_ref"), P = /* @__PURE__ */ Symbol.for("react.suspense"), $t = /* @__PURE__ */ Symbol.for("react.suspense_list"), V = /* @__PURE__ */ Symbol.for("react.memo"), zt = /* @__PURE__ */ Symbol.for("react.lazy"), Wt = /* @__PURE__ */ Symbol.for("react.activity"), Fl = /* @__PURE__ */ Symbol.for("react.memo_cache_sentinel"), nl = Symbol.iterator;
  function Zt(t) {
    return t === null || typeof t != "object" ? null : (t = nl && t[nl] || t["@@iterator"], typeof t == "function" ? t : null);
  }
  var Ql = /* @__PURE__ */ Symbol.for("react.client.reference");
  function vl(t) {
    if (t == null) return null;
    if (typeof t == "function")
      return t.$$typeof === Ql ? null : t.displayName || t.name || null;
    if (typeof t == "string") return t;
    switch (t) {
      case B:
        return "Fragment";
      case Ht:
        return "Profiler";
      case rt:
        return "StrictMode";
      case P:
        return "Suspense";
      case $t:
        return "SuspenseList";
      case Wt:
        return "Activity";
    }
    if (typeof t == "object")
      switch (t.$$typeof) {
        case st:
          return "Portal";
        case Et:
          return (t.displayName || "Context") + ".Provider";
        case qt:
          return (t._context.displayName || "Context") + ".Consumer";
        case nt:
          var l = t.render;
          return t = t.displayName, t || (t = l.displayName || l.name || "", t = t !== "" ? "ForwardRef(" + t + ")" : "ForwardRef"), t;
        case V:
          return l = t.displayName || null, l !== null ? l : vl(t.type) || "Memo";
        case zt:
          l = t._payload, t = t._init;
          try {
            return vl(t(l));
          } catch {
          }
      }
    return null;
  }
  var Dt = Array.isArray, p = x.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, H = A.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, w = {
    pending: !1,
    data: null,
    method: null,
    action: null
  }, ot = [], d = -1;
  function M(t) {
    return { current: t };
  }
  function Y(t) {
    0 > d || (t.current = ot[d], ot[d] = null, d--);
  }
  function q(t, l) {
    d++, ot[d] = t.current, t.current = l;
  }
  var L = M(null), it = M(null), $ = M(null), ll = M(null);
  function Tt(t, l) {
    switch (q($, l), q(it, t), q(L, null), l.nodeType) {
      case 9:
      case 11:
        t = (t = l.documentElement) && (t = t.namespaceURI) ? Fo(t) : 0;
        break;
      default:
        if (t = l.tagName, l = l.namespaceURI)
          l = Fo(l), t = Po(l, t);
        else
          switch (t) {
            case "svg":
              t = 1;
              break;
            case "math":
              t = 2;
              break;
            default:
              t = 0;
          }
    }
    Y(L), q(L, t);
  }
  function Zl() {
    Y(L), Y(it), Y($);
  }
  function Pa(t) {
    t.memoizedState !== null && q(ll, t);
    var l = L.current, e = Po(l, t.type);
    l !== e && (q(it, t), q(L, e));
  }
  function Se(t) {
    it.current === t && (Y(L), Y(it)), ll.current === t && (Y(ll), Lu._currentValue = w);
  }
  var ml = Object.prototype.hasOwnProperty, oa = r.unstable_scheduleCallback, Pl = r.unstable_cancelCallback, Fu = r.unstable_shouldYield, yi = r.unstable_requestPaint, gl = r.unstable_now, Sl = r.unstable_getCurrentPriorityLevel, Ze = r.unstable_ImmediatePriority, Il = r.unstable_UserBlockingPriority, Ml = r.unstable_NormalPriority, bl = r.unstable_LowPriority, be = r.unstable_IdlePriority, Ll = r.log, _t = r.unstable_setDisableYieldValue, Lt = null, Ft = null;
  function il(t) {
    if (typeof Ll == "function" && _t(t), Ft && typeof Ft.setStrictMode == "function")
      try {
        Ft.setStrictMode(Lt, t);
      } catch {
      }
  }
  var ct = Math.clz32 ? Math.clz32 : mi, Pu = Math.log, vi = Math.LN2;
  function mi(t) {
    return t >>>= 0, t === 0 ? 32 : 31 - (Pu(t) / vi | 0) | 0;
  }
  var Le = 256, Ve = 4194304;
  function te(t) {
    var l = t & 42;
    if (l !== 0) return l;
    switch (t & -t) {
      case 1:
        return 1;
      case 2:
        return 2;
      case 4:
        return 4;
      case 8:
        return 8;
      case 16:
        return 16;
      case 32:
        return 32;
      case 64:
        return 64;
      case 128:
        return 128;
      case 256:
      case 512:
      case 1024:
      case 2048:
      case 4096:
      case 8192:
      case 16384:
      case 32768:
      case 65536:
      case 131072:
      case 262144:
      case 524288:
      case 1048576:
      case 2097152:
        return t & 4194048;
      case 4194304:
      case 8388608:
      case 16777216:
      case 33554432:
        return t & 62914560;
      case 67108864:
        return 67108864;
      case 134217728:
        return 134217728;
      case 268435456:
        return 268435456;
      case 536870912:
        return 536870912;
      case 1073741824:
        return 0;
      default:
        return t;
    }
  }
  function da(t, l, e) {
    var a = t.pendingLanes;
    if (a === 0) return 0;
    var u = 0, n = t.suspendedLanes, i = t.pingedLanes;
    t = t.warmLanes;
    var f = a & 134217727;
    return f !== 0 ? (a = f & ~n, a !== 0 ? u = te(a) : (i &= f, i !== 0 ? u = te(i) : e || (e = f & ~t, e !== 0 && (u = te(e))))) : (f = a & ~n, f !== 0 ? u = te(f) : i !== 0 ? u = te(i) : e || (e = a & ~t, e !== 0 && (u = te(e)))), u === 0 ? 0 : l !== 0 && l !== u && (l & n) === 0 && (n = u & -u, e = l & -l, n >= e || n === 32 && (e & 4194048) !== 0) ? l : u;
  }
  function Vl(t, l) {
    return (t.pendingLanes & ~(t.suspendedLanes & ~t.pingedLanes) & l) === 0;
  }
  function we(t, l) {
    switch (t) {
      case 1:
      case 2:
      case 4:
      case 8:
      case 64:
        return l + 250;
      case 16:
      case 32:
      case 128:
      case 256:
      case 512:
      case 1024:
      case 2048:
      case 4096:
      case 8192:
      case 16384:
      case 32768:
      case 65536:
      case 131072:
      case 262144:
      case 524288:
      case 1048576:
      case 2097152:
        return l + 5e3;
      case 4194304:
      case 8388608:
      case 16777216:
      case 33554432:
        return -1;
      case 67108864:
      case 134217728:
      case 268435456:
      case 536870912:
      case 1073741824:
        return -1;
      default:
        return -1;
    }
  }
  function E() {
    var t = Le;
    return Le <<= 1, (Le & 4194048) === 0 && (Le = 256), t;
  }
  function X() {
    var t = Ve;
    return Ve <<= 1, (Ve & 62914560) === 0 && (Ve = 4194304), t;
  }
  function F(t) {
    for (var l = [], e = 0; 31 > e; e++) l.push(t);
    return l;
  }
  function tt(t, l) {
    t.pendingLanes |= l, l !== 268435456 && (t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0);
  }
  function dt(t, l, e, a, u, n) {
    var i = t.pendingLanes;
    t.pendingLanes = e, t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0, t.expiredLanes &= e, t.entangledLanes &= e, t.errorRecoveryDisabledLanes &= e, t.shellSuspendCounter = 0;
    var f = t.entanglements, s = t.expirationTimes, m = t.hiddenUpdates;
    for (e = i & ~e; 0 < e; ) {
      var j = 31 - ct(e), z = 1 << j;
      f[j] = 0, s[j] = -1;
      var g = m[j];
      if (g !== null)
        for (m[j] = null, j = 0; j < g.length; j++) {
          var S = g[j];
          S !== null && (S.lane &= -536870913);
        }
      e &= ~z;
    }
    a !== 0 && Vt(t, a, 0), n !== 0 && u === 0 && t.tag !== 0 && (t.suspendedLanes |= n & ~(i & ~l));
  }
  function Vt(t, l, e) {
    t.pendingLanes |= l, t.suspendedLanes &= ~l;
    var a = 31 - ct(l);
    t.entangledLanes |= l, t.entanglements[a] = t.entanglements[a] | 1073741824 | e & 4194090;
  }
  function Ke(t, l) {
    var e = t.entangledLanes |= l;
    for (t = t.entanglements; e; ) {
      var a = 31 - ct(e), u = 1 << a;
      u & l | t[a] & l && (t[a] |= l), e &= ~u;
    }
  }
  function cl(t) {
    switch (t) {
      case 2:
        t = 1;
        break;
      case 8:
        t = 4;
        break;
      case 32:
        t = 16;
        break;
      case 256:
      case 512:
      case 1024:
      case 2048:
      case 4096:
      case 8192:
      case 16384:
      case 32768:
      case 65536:
      case 131072:
      case 262144:
      case 524288:
      case 1048576:
      case 2097152:
      case 4194304:
      case 8388608:
      case 16777216:
      case 33554432:
        t = 128;
        break;
      case 268435456:
        t = 134217728;
        break;
      default:
        t = 0;
    }
    return t;
  }
  function Bl(t) {
    return t &= -t, 2 < t ? 8 < t ? (t & 134217727) !== 0 ? 32 : 268435456 : 8 : 2;
  }
  function le() {
    var t = H.p;
    return t !== 0 ? t : (t = window.event, t === void 0 ? 32 : gd(t.type));
  }
  function Ia(t, l) {
    var e = H.p;
    try {
      return H.p = t, l();
    } finally {
      H.p = e;
    }
  }
  var wl = Math.random().toString(36).slice(2), el = "__reactFiber$" + wl, sl = "__reactProps$" + wl, ha = "__reactContainer$" + wl, gi = "__reactEvents$" + wl, Ld = "__reactListeners$" + wl, Vd = "__reactHandles$" + wl, Yf = "__reactResources$" + wl, tu = "__reactMarker$" + wl;
  function Si(t) {
    delete t[el], delete t[sl], delete t[gi], delete t[Ld], delete t[Vd];
  }
  function ya(t) {
    var l = t[el];
    if (l) return l;
    for (var e = t.parentNode; e; ) {
      if (l = e[ha] || e[el]) {
        if (e = l.alternate, l.child !== null || e !== null && e.child !== null)
          for (t = ed(t); t !== null; ) {
            if (e = t[el]) return e;
            t = ed(t);
          }
        return l;
      }
      t = e, e = t.parentNode;
    }
    return null;
  }
  function va(t) {
    if (t = t[el] || t[ha]) {
      var l = t.tag;
      if (l === 5 || l === 6 || l === 13 || l === 26 || l === 27 || l === 3)
        return t;
    }
    return null;
  }
  function lu(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t.stateNode;
    throw Error(o(33));
  }
  function ma(t) {
    var l = t[Yf];
    return l || (l = t[Yf] = { hoistableStyles: /* @__PURE__ */ new Map(), hoistableScripts: /* @__PURE__ */ new Map() }), l;
  }
  function wt(t) {
    t[tu] = !0;
  }
  var Gf = /* @__PURE__ */ new Set(), Xf = {};
  function Je(t, l) {
    ga(t, l), ga(t + "Capture", l);
  }
  function ga(t, l) {
    for (Xf[t] = l, t = 0; t < l.length; t++)
      Gf.add(l[t]);
  }
  var wd = RegExp(
    "^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"
  ), Qf = {}, Zf = {};
  function Kd(t) {
    return ml.call(Zf, t) ? !0 : ml.call(Qf, t) ? !1 : wd.test(t) ? Zf[t] = !0 : (Qf[t] = !0, !1);
  }
  function Iu(t, l, e) {
    if (Kd(l))
      if (e === null) t.removeAttribute(l);
      else {
        switch (typeof e) {
          case "undefined":
          case "function":
          case "symbol":
            t.removeAttribute(l);
            return;
          case "boolean":
            var a = l.toLowerCase().slice(0, 5);
            if (a !== "data-" && a !== "aria-") {
              t.removeAttribute(l);
              return;
            }
        }
        t.setAttribute(l, "" + e);
      }
  }
  function tn(t, l, e) {
    if (e === null) t.removeAttribute(l);
    else {
      switch (typeof e) {
        case "undefined":
        case "function":
        case "symbol":
        case "boolean":
          t.removeAttribute(l);
          return;
      }
      t.setAttribute(l, "" + e);
    }
  }
  function ee(t, l, e, a) {
    if (a === null) t.removeAttribute(e);
    else {
      switch (typeof a) {
        case "undefined":
        case "function":
        case "symbol":
        case "boolean":
          t.removeAttribute(e);
          return;
      }
      t.setAttributeNS(l, e, "" + a);
    }
  }
  var bi, Lf;
  function Sa(t) {
    if (bi === void 0)
      try {
        throw Error();
      } catch (e) {
        var l = e.stack.trim().match(/\n( *(at )?)/);
        bi = l && l[1] || "", Lf = -1 < e.stack.indexOf(`
    at`) ? " (<anonymous>)" : -1 < e.stack.indexOf("@") ? "@unknown:0:0" : "";
      }
    return `
` + bi + t + Lf;
  }
  var xi = !1;
  function pi(t, l) {
    if (!t || xi) return "";
    xi = !0;
    var e = Error.prepareStackTrace;
    Error.prepareStackTrace = void 0;
    try {
      var a = {
        DetermineComponentFrameRoot: function() {
          try {
            if (l) {
              var z = function() {
                throw Error();
              };
              if (Object.defineProperty(z.prototype, "props", {
                set: function() {
                  throw Error();
                }
              }), typeof Reflect == "object" && Reflect.construct) {
                try {
                  Reflect.construct(z, []);
                } catch (S) {
                  var g = S;
                }
                Reflect.construct(t, [], z);
              } else {
                try {
                  z.call();
                } catch (S) {
                  g = S;
                }
                t.call(z.prototype);
              }
            } else {
              try {
                throw Error();
              } catch (S) {
                g = S;
              }
              (z = t()) && typeof z.catch == "function" && z.catch(function() {
              });
            }
          } catch (S) {
            if (S && g && typeof S.stack == "string")
              return [S.stack, g.stack];
          }
          return [null, null];
        }
      };
      a.DetermineComponentFrameRoot.displayName = "DetermineComponentFrameRoot";
      var u = Object.getOwnPropertyDescriptor(
        a.DetermineComponentFrameRoot,
        "name"
      );
      u && u.configurable && Object.defineProperty(
        a.DetermineComponentFrameRoot,
        "name",
        { value: "DetermineComponentFrameRoot" }
      );
      var n = a.DetermineComponentFrameRoot(), i = n[0], f = n[1];
      if (i && f) {
        var s = i.split(`
`), m = f.split(`
`);
        for (u = a = 0; a < s.length && !s[a].includes("DetermineComponentFrameRoot"); )
          a++;
        for (; u < m.length && !m[u].includes(
          "DetermineComponentFrameRoot"
        ); )
          u++;
        if (a === s.length || u === m.length)
          for (a = s.length - 1, u = m.length - 1; 1 <= a && 0 <= u && s[a] !== m[u]; )
            u--;
        for (; 1 <= a && 0 <= u; a--, u--)
          if (s[a] !== m[u]) {
            if (a !== 1 || u !== 1)
              do
                if (a--, u--, 0 > u || s[a] !== m[u]) {
                  var j = `
` + s[a].replace(" at new ", " at ");
                  return t.displayName && j.includes("<anonymous>") && (j = j.replace("<anonymous>", t.displayName)), j;
                }
              while (1 <= a && 0 <= u);
            break;
          }
      }
    } finally {
      xi = !1, Error.prepareStackTrace = e;
    }
    return (e = t ? t.displayName || t.name : "") ? Sa(e) : "";
  }
  function Jd(t) {
    switch (t.tag) {
      case 26:
      case 27:
      case 5:
        return Sa(t.type);
      case 16:
        return Sa("Lazy");
      case 13:
        return Sa("Suspense");
      case 19:
        return Sa("SuspenseList");
      case 0:
      case 15:
        return pi(t.type, !1);
      case 11:
        return pi(t.type.render, !1);
      case 1:
        return pi(t.type, !0);
      case 31:
        return Sa("Activity");
      default:
        return "";
    }
  }
  function Vf(t) {
    try {
      var l = "";
      do
        l += Jd(t), t = t.return;
      while (t);
      return l;
    } catch (e) {
      return `
Error generating stack: ` + e.message + `
` + e.stack;
    }
  }
  function Nl(t) {
    switch (typeof t) {
      case "bigint":
      case "boolean":
      case "number":
      case "string":
      case "undefined":
        return t;
      case "object":
        return t;
      default:
        return "";
    }
  }
  function wf(t) {
    var l = t.type;
    return (t = t.nodeName) && t.toLowerCase() === "input" && (l === "checkbox" || l === "radio");
  }
  function kd(t) {
    var l = wf(t) ? "checked" : "value", e = Object.getOwnPropertyDescriptor(
      t.constructor.prototype,
      l
    ), a = "" + t[l];
    if (!t.hasOwnProperty(l) && typeof e < "u" && typeof e.get == "function" && typeof e.set == "function") {
      var u = e.get, n = e.set;
      return Object.defineProperty(t, l, {
        configurable: !0,
        get: function() {
          return u.call(this);
        },
        set: function(i) {
          a = "" + i, n.call(this, i);
        }
      }), Object.defineProperty(t, l, {
        enumerable: e.enumerable
      }), {
        getValue: function() {
          return a;
        },
        setValue: function(i) {
          a = "" + i;
        },
        stopTracking: function() {
          t._valueTracker = null, delete t[l];
        }
      };
    }
  }
  function ln(t) {
    t._valueTracker || (t._valueTracker = kd(t));
  }
  function Kf(t) {
    if (!t) return !1;
    var l = t._valueTracker;
    if (!l) return !0;
    var e = l.getValue(), a = "";
    return t && (a = wf(t) ? t.checked ? "true" : "false" : t.value), t = a, t !== e ? (l.setValue(t), !0) : !1;
  }
  function en(t) {
    if (t = t || (typeof document < "u" ? document : void 0), typeof t > "u") return null;
    try {
      return t.activeElement || t.body;
    } catch {
      return t.body;
    }
  }
  var $d = /[\n"\\]/g;
  function Dl(t) {
    return t.replace(
      $d,
      function(l) {
        return "\\" + l.charCodeAt(0).toString(16) + " ";
      }
    );
  }
  function ji(t, l, e, a, u, n, i, f) {
    t.name = "", i != null && typeof i != "function" && typeof i != "symbol" && typeof i != "boolean" ? t.type = i : t.removeAttribute("type"), l != null ? i === "number" ? (l === 0 && t.value === "" || t.value != l) && (t.value = "" + Nl(l)) : t.value !== "" + Nl(l) && (t.value = "" + Nl(l)) : i !== "submit" && i !== "reset" || t.removeAttribute("value"), l != null ? Ti(t, i, Nl(l)) : e != null ? Ti(t, i, Nl(e)) : a != null && t.removeAttribute("value"), u == null && n != null && (t.defaultChecked = !!n), u != null && (t.checked = u && typeof u != "function" && typeof u != "symbol"), f != null && typeof f != "function" && typeof f != "symbol" && typeof f != "boolean" ? t.name = "" + Nl(f) : t.removeAttribute("name");
  }
  function Jf(t, l, e, a, u, n, i, f) {
    if (n != null && typeof n != "function" && typeof n != "symbol" && typeof n != "boolean" && (t.type = n), l != null || e != null) {
      if (!(n !== "submit" && n !== "reset" || l != null))
        return;
      e = e != null ? "" + Nl(e) : "", l = l != null ? "" + Nl(l) : e, f || l === t.value || (t.value = l), t.defaultValue = l;
    }
    a = a ?? u, a = typeof a != "function" && typeof a != "symbol" && !!a, t.checked = f ? t.checked : !!a, t.defaultChecked = !!a, i != null && typeof i != "function" && typeof i != "symbol" && typeof i != "boolean" && (t.name = i);
  }
  function Ti(t, l, e) {
    l === "number" && en(t.ownerDocument) === t || t.defaultValue === "" + e || (t.defaultValue = "" + e);
  }
  function ba(t, l, e, a) {
    if (t = t.options, l) {
      l = {};
      for (var u = 0; u < e.length; u++)
        l["$" + e[u]] = !0;
      for (e = 0; e < t.length; e++)
        u = l.hasOwnProperty("$" + t[e].value), t[e].selected !== u && (t[e].selected = u), u && a && (t[e].defaultSelected = !0);
    } else {
      for (e = "" + Nl(e), l = null, u = 0; u < t.length; u++) {
        if (t[u].value === e) {
          t[u].selected = !0, a && (t[u].defaultSelected = !0);
          return;
        }
        l !== null || t[u].disabled || (l = t[u]);
      }
      l !== null && (l.selected = !0);
    }
  }
  function kf(t, l, e) {
    if (l != null && (l = "" + Nl(l), l !== t.value && (t.value = l), e == null)) {
      t.defaultValue !== l && (t.defaultValue = l);
      return;
    }
    t.defaultValue = e != null ? "" + Nl(e) : "";
  }
  function $f(t, l, e, a) {
    if (l == null) {
      if (a != null) {
        if (e != null) throw Error(o(92));
        if (Dt(a)) {
          if (1 < a.length) throw Error(o(93));
          a = a[0];
        }
        e = a;
      }
      e == null && (e = ""), l = e;
    }
    e = Nl(l), t.defaultValue = e, a = t.textContent, a === e && a !== "" && a !== null && (t.value = a);
  }
  function xa(t, l) {
    if (l) {
      var e = t.firstChild;
      if (e && e === t.lastChild && e.nodeType === 3) {
        e.nodeValue = l;
        return;
      }
    }
    t.textContent = l;
  }
  var Wd = new Set(
    "animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(
      " "
    )
  );
  function Wf(t, l, e) {
    var a = l.indexOf("--") === 0;
    e == null || typeof e == "boolean" || e === "" ? a ? t.setProperty(l, "") : l === "float" ? t.cssFloat = "" : t[l] = "" : a ? t.setProperty(l, e) : typeof e != "number" || e === 0 || Wd.has(l) ? l === "float" ? t.cssFloat = e : t[l] = ("" + e).trim() : t[l] = e + "px";
  }
  function Ff(t, l, e) {
    if (l != null && typeof l != "object")
      throw Error(o(62));
    if (t = t.style, e != null) {
      for (var a in e)
        !e.hasOwnProperty(a) || l != null && l.hasOwnProperty(a) || (a.indexOf("--") === 0 ? t.setProperty(a, "") : a === "float" ? t.cssFloat = "" : t[a] = "");
      for (var u in l)
        a = l[u], l.hasOwnProperty(u) && e[u] !== a && Wf(t, u, a);
    } else
      for (var n in l)
        l.hasOwnProperty(n) && Wf(t, n, l[n]);
  }
  function Ai(t) {
    if (t.indexOf("-") === -1) return !1;
    switch (t) {
      case "annotation-xml":
      case "color-profile":
      case "font-face":
      case "font-face-src":
      case "font-face-uri":
      case "font-face-format":
      case "font-face-name":
      case "missing-glyph":
        return !1;
      default:
        return !0;
    }
  }
  var Fd = /* @__PURE__ */ new Map([
    ["acceptCharset", "accept-charset"],
    ["htmlFor", "for"],
    ["httpEquiv", "http-equiv"],
    ["crossOrigin", "crossorigin"],
    ["accentHeight", "accent-height"],
    ["alignmentBaseline", "alignment-baseline"],
    ["arabicForm", "arabic-form"],
    ["baselineShift", "baseline-shift"],
    ["capHeight", "cap-height"],
    ["clipPath", "clip-path"],
    ["clipRule", "clip-rule"],
    ["colorInterpolation", "color-interpolation"],
    ["colorInterpolationFilters", "color-interpolation-filters"],
    ["colorProfile", "color-profile"],
    ["colorRendering", "color-rendering"],
    ["dominantBaseline", "dominant-baseline"],
    ["enableBackground", "enable-background"],
    ["fillOpacity", "fill-opacity"],
    ["fillRule", "fill-rule"],
    ["floodColor", "flood-color"],
    ["floodOpacity", "flood-opacity"],
    ["fontFamily", "font-family"],
    ["fontSize", "font-size"],
    ["fontSizeAdjust", "font-size-adjust"],
    ["fontStretch", "font-stretch"],
    ["fontStyle", "font-style"],
    ["fontVariant", "font-variant"],
    ["fontWeight", "font-weight"],
    ["glyphName", "glyph-name"],
    ["glyphOrientationHorizontal", "glyph-orientation-horizontal"],
    ["glyphOrientationVertical", "glyph-orientation-vertical"],
    ["horizAdvX", "horiz-adv-x"],
    ["horizOriginX", "horiz-origin-x"],
    ["imageRendering", "image-rendering"],
    ["letterSpacing", "letter-spacing"],
    ["lightingColor", "lighting-color"],
    ["markerEnd", "marker-end"],
    ["markerMid", "marker-mid"],
    ["markerStart", "marker-start"],
    ["overlinePosition", "overline-position"],
    ["overlineThickness", "overline-thickness"],
    ["paintOrder", "paint-order"],
    ["panose-1", "panose-1"],
    ["pointerEvents", "pointer-events"],
    ["renderingIntent", "rendering-intent"],
    ["shapeRendering", "shape-rendering"],
    ["stopColor", "stop-color"],
    ["stopOpacity", "stop-opacity"],
    ["strikethroughPosition", "strikethrough-position"],
    ["strikethroughThickness", "strikethrough-thickness"],
    ["strokeDasharray", "stroke-dasharray"],
    ["strokeDashoffset", "stroke-dashoffset"],
    ["strokeLinecap", "stroke-linecap"],
    ["strokeLinejoin", "stroke-linejoin"],
    ["strokeMiterlimit", "stroke-miterlimit"],
    ["strokeOpacity", "stroke-opacity"],
    ["strokeWidth", "stroke-width"],
    ["textAnchor", "text-anchor"],
    ["textDecoration", "text-decoration"],
    ["textRendering", "text-rendering"],
    ["transformOrigin", "transform-origin"],
    ["underlinePosition", "underline-position"],
    ["underlineThickness", "underline-thickness"],
    ["unicodeBidi", "unicode-bidi"],
    ["unicodeRange", "unicode-range"],
    ["unitsPerEm", "units-per-em"],
    ["vAlphabetic", "v-alphabetic"],
    ["vHanging", "v-hanging"],
    ["vIdeographic", "v-ideographic"],
    ["vMathematical", "v-mathematical"],
    ["vectorEffect", "vector-effect"],
    ["vertAdvY", "vert-adv-y"],
    ["vertOriginX", "vert-origin-x"],
    ["vertOriginY", "vert-origin-y"],
    ["wordSpacing", "word-spacing"],
    ["writingMode", "writing-mode"],
    ["xmlnsXlink", "xmlns:xlink"],
    ["xHeight", "x-height"]
  ]), Pd = /^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;
  function an(t) {
    return Pd.test("" + t) ? "javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')" : t;
  }
  var Ei = null;
  function zi(t) {
    return t = t.target || t.srcElement || window, t.correspondingUseElement && (t = t.correspondingUseElement), t.nodeType === 3 ? t.parentNode : t;
  }
  var pa = null, ja = null;
  function Pf(t) {
    var l = va(t);
    if (l && (t = l.stateNode)) {
      var e = t[sl] || null;
      t: switch (t = l.stateNode, l.type) {
        case "input":
          if (ji(
            t,
            e.value,
            e.defaultValue,
            e.defaultValue,
            e.checked,
            e.defaultChecked,
            e.type,
            e.name
          ), l = e.name, e.type === "radio" && l != null) {
            for (e = t; e.parentNode; ) e = e.parentNode;
            for (e = e.querySelectorAll(
              'input[name="' + Dl(
                "" + l
              ) + '"][type="radio"]'
            ), l = 0; l < e.length; l++) {
              var a = e[l];
              if (a !== t && a.form === t.form) {
                var u = a[sl] || null;
                if (!u) throw Error(o(90));
                ji(
                  a,
                  u.value,
                  u.defaultValue,
                  u.defaultValue,
                  u.checked,
                  u.defaultChecked,
                  u.type,
                  u.name
                );
              }
            }
            for (l = 0; l < e.length; l++)
              a = e[l], a.form === t.form && Kf(a);
          }
          break t;
        case "textarea":
          kf(t, e.value, e.defaultValue);
          break t;
        case "select":
          l = e.value, l != null && ba(t, !!e.multiple, l, !1);
      }
    }
  }
  var Mi = !1;
  function If(t, l, e) {
    if (Mi) return t(l, e);
    Mi = !0;
    try {
      var a = t(l);
      return a;
    } finally {
      if (Mi = !1, (pa !== null || ja !== null) && (Ln(), pa && (l = pa, t = ja, ja = pa = null, Pf(l), t)))
        for (l = 0; l < t.length; l++) Pf(t[l]);
    }
  }
  function eu(t, l) {
    var e = t.stateNode;
    if (e === null) return null;
    var a = e[sl] || null;
    if (a === null) return null;
    e = a[l];
    t: switch (l) {
      case "onClick":
      case "onClickCapture":
      case "onDoubleClick":
      case "onDoubleClickCapture":
      case "onMouseDown":
      case "onMouseDownCapture":
      case "onMouseMove":
      case "onMouseMoveCapture":
      case "onMouseUp":
      case "onMouseUpCapture":
      case "onMouseEnter":
        (a = !a.disabled) || (t = t.type, a = !(t === "button" || t === "input" || t === "select" || t === "textarea")), t = !a;
        break t;
      default:
        t = !1;
    }
    if (t) return null;
    if (e && typeof e != "function")
      throw Error(
        o(231, l, typeof e)
      );
    return e;
  }
  var ae = !(typeof window > "u" || typeof window.document > "u" || typeof window.document.createElement > "u"), Ni = !1;
  if (ae)
    try {
      var au = {};
      Object.defineProperty(au, "passive", {
        get: function() {
          Ni = !0;
        }
      }), window.addEventListener("test", au, au), window.removeEventListener("test", au, au);
    } catch {
      Ni = !1;
    }
  var xe = null, Di = null, un = null;
  function ts() {
    if (un) return un;
    var t, l = Di, e = l.length, a, u = "value" in xe ? xe.value : xe.textContent, n = u.length;
    for (t = 0; t < e && l[t] === u[t]; t++) ;
    var i = e - t;
    for (a = 1; a <= i && l[e - a] === u[n - a]; a++) ;
    return un = u.slice(t, 1 < a ? 1 - a : void 0);
  }
  function nn(t) {
    var l = t.keyCode;
    return "charCode" in t ? (t = t.charCode, t === 0 && l === 13 && (t = 13)) : t = l, t === 10 && (t = 13), 32 <= t || t === 13 ? t : 0;
  }
  function cn() {
    return !0;
  }
  function ls() {
    return !1;
  }
  function rl(t) {
    function l(e, a, u, n, i) {
      this._reactName = e, this._targetInst = u, this.type = a, this.nativeEvent = n, this.target = i, this.currentTarget = null;
      for (var f in t)
        t.hasOwnProperty(f) && (e = t[f], this[f] = e ? e(n) : n[f]);
      return this.isDefaultPrevented = (n.defaultPrevented != null ? n.defaultPrevented : n.returnValue === !1) ? cn : ls, this.isPropagationStopped = ls, this;
    }
    return C(l.prototype, {
      preventDefault: function() {
        this.defaultPrevented = !0;
        var e = this.nativeEvent;
        e && (e.preventDefault ? e.preventDefault() : typeof e.returnValue != "unknown" && (e.returnValue = !1), this.isDefaultPrevented = cn);
      },
      stopPropagation: function() {
        var e = this.nativeEvent;
        e && (e.stopPropagation ? e.stopPropagation() : typeof e.cancelBubble != "unknown" && (e.cancelBubble = !0), this.isPropagationStopped = cn);
      },
      persist: function() {
      },
      isPersistent: cn
    }), l;
  }
  var ke = {
    eventPhase: 0,
    bubbles: 0,
    cancelable: 0,
    timeStamp: function(t) {
      return t.timeStamp || Date.now();
    },
    defaultPrevented: 0,
    isTrusted: 0
  }, fn = rl(ke), uu = C({}, ke, { view: 0, detail: 0 }), Id = rl(uu), Oi, Ri, nu, sn = C({}, uu, {
    screenX: 0,
    screenY: 0,
    clientX: 0,
    clientY: 0,
    pageX: 0,
    pageY: 0,
    ctrlKey: 0,
    shiftKey: 0,
    altKey: 0,
    metaKey: 0,
    getModifierState: Ui,
    button: 0,
    buttons: 0,
    relatedTarget: function(t) {
      return t.relatedTarget === void 0 ? t.fromElement === t.srcElement ? t.toElement : t.fromElement : t.relatedTarget;
    },
    movementX: function(t) {
      return "movementX" in t ? t.movementX : (t !== nu && (nu && t.type === "mousemove" ? (Oi = t.screenX - nu.screenX, Ri = t.screenY - nu.screenY) : Ri = Oi = 0, nu = t), Oi);
    },
    movementY: function(t) {
      return "movementY" in t ? t.movementY : Ri;
    }
  }), es = rl(sn), th = C({}, sn, { dataTransfer: 0 }), lh = rl(th), eh = C({}, uu, { relatedTarget: 0 }), _i = rl(eh), ah = C({}, ke, {
    animationName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), uh = rl(ah), nh = C({}, ke, {
    clipboardData: function(t) {
      return "clipboardData" in t ? t.clipboardData : window.clipboardData;
    }
  }), ih = rl(nh), ch = C({}, ke, { data: 0 }), as = rl(ch), fh = {
    Esc: "Escape",
    Spacebar: " ",
    Left: "ArrowLeft",
    Up: "ArrowUp",
    Right: "ArrowRight",
    Down: "ArrowDown",
    Del: "Delete",
    Win: "OS",
    Menu: "ContextMenu",
    Apps: "ContextMenu",
    Scroll: "ScrollLock",
    MozPrintableKey: "Unidentified"
  }, sh = {
    8: "Backspace",
    9: "Tab",
    12: "Clear",
    13: "Enter",
    16: "Shift",
    17: "Control",
    18: "Alt",
    19: "Pause",
    20: "CapsLock",
    27: "Escape",
    32: " ",
    33: "PageUp",
    34: "PageDown",
    35: "End",
    36: "Home",
    37: "ArrowLeft",
    38: "ArrowUp",
    39: "ArrowRight",
    40: "ArrowDown",
    45: "Insert",
    46: "Delete",
    112: "F1",
    113: "F2",
    114: "F3",
    115: "F4",
    116: "F5",
    117: "F6",
    118: "F7",
    119: "F8",
    120: "F9",
    121: "F10",
    122: "F11",
    123: "F12",
    144: "NumLock",
    145: "ScrollLock",
    224: "Meta"
  }, rh = {
    Alt: "altKey",
    Control: "ctrlKey",
    Meta: "metaKey",
    Shift: "shiftKey"
  };
  function oh(t) {
    var l = this.nativeEvent;
    return l.getModifierState ? l.getModifierState(t) : (t = rh[t]) ? !!l[t] : !1;
  }
  function Ui() {
    return oh;
  }
  var dh = C({}, uu, {
    key: function(t) {
      if (t.key) {
        var l = fh[t.key] || t.key;
        if (l !== "Unidentified") return l;
      }
      return t.type === "keypress" ? (t = nn(t), t === 13 ? "Enter" : String.fromCharCode(t)) : t.type === "keydown" || t.type === "keyup" ? sh[t.keyCode] || "Unidentified" : "";
    },
    code: 0,
    location: 0,
    ctrlKey: 0,
    shiftKey: 0,
    altKey: 0,
    metaKey: 0,
    repeat: 0,
    locale: 0,
    getModifierState: Ui,
    charCode: function(t) {
      return t.type === "keypress" ? nn(t) : 0;
    },
    keyCode: function(t) {
      return t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    },
    which: function(t) {
      return t.type === "keypress" ? nn(t) : t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    }
  }), hh = rl(dh), yh = C({}, sn, {
    pointerId: 0,
    width: 0,
    height: 0,
    pressure: 0,
    tangentialPressure: 0,
    tiltX: 0,
    tiltY: 0,
    twist: 0,
    pointerType: 0,
    isPrimary: 0
  }), us = rl(yh), vh = C({}, uu, {
    touches: 0,
    targetTouches: 0,
    changedTouches: 0,
    altKey: 0,
    metaKey: 0,
    ctrlKey: 0,
    shiftKey: 0,
    getModifierState: Ui
  }), mh = rl(vh), gh = C({}, ke, {
    propertyName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), Sh = rl(gh), bh = C({}, sn, {
    deltaX: function(t) {
      return "deltaX" in t ? t.deltaX : "wheelDeltaX" in t ? -t.wheelDeltaX : 0;
    },
    deltaY: function(t) {
      return "deltaY" in t ? t.deltaY : "wheelDeltaY" in t ? -t.wheelDeltaY : "wheelDelta" in t ? -t.wheelDelta : 0;
    },
    deltaZ: 0,
    deltaMode: 0
  }), xh = rl(bh), ph = C({}, ke, {
    newState: 0,
    oldState: 0
  }), jh = rl(ph), Th = [9, 13, 27, 32], Ci = ae && "CompositionEvent" in window, iu = null;
  ae && "documentMode" in document && (iu = document.documentMode);
  var Ah = ae && "TextEvent" in window && !iu, ns = ae && (!Ci || iu && 8 < iu && 11 >= iu), is = " ", cs = !1;
  function fs(t, l) {
    switch (t) {
      case "keyup":
        return Th.indexOf(l.keyCode) !== -1;
      case "keydown":
        return l.keyCode !== 229;
      case "keypress":
      case "mousedown":
      case "focusout":
        return !0;
      default:
        return !1;
    }
  }
  function ss(t) {
    return t = t.detail, typeof t == "object" && "data" in t ? t.data : null;
  }
  var Ta = !1;
  function Eh(t, l) {
    switch (t) {
      case "compositionend":
        return ss(l);
      case "keypress":
        return l.which !== 32 ? null : (cs = !0, is);
      case "textInput":
        return t = l.data, t === is && cs ? null : t;
      default:
        return null;
    }
  }
  function zh(t, l) {
    if (Ta)
      return t === "compositionend" || !Ci && fs(t, l) ? (t = ts(), un = Di = xe = null, Ta = !1, t) : null;
    switch (t) {
      case "paste":
        return null;
      case "keypress":
        if (!(l.ctrlKey || l.altKey || l.metaKey) || l.ctrlKey && l.altKey) {
          if (l.char && 1 < l.char.length)
            return l.char;
          if (l.which) return String.fromCharCode(l.which);
        }
        return null;
      case "compositionend":
        return ns && l.locale !== "ko" ? null : l.data;
      default:
        return null;
    }
  }
  var Mh = {
    color: !0,
    date: !0,
    datetime: !0,
    "datetime-local": !0,
    email: !0,
    month: !0,
    number: !0,
    password: !0,
    range: !0,
    search: !0,
    tel: !0,
    text: !0,
    time: !0,
    url: !0,
    week: !0
  };
  function rs(t) {
    var l = t && t.nodeName && t.nodeName.toLowerCase();
    return l === "input" ? !!Mh[t.type] : l === "textarea";
  }
  function os(t, l, e, a) {
    pa ? ja ? ja.push(a) : ja = [a] : pa = a, l = $n(l, "onChange"), 0 < l.length && (e = new fn(
      "onChange",
      "change",
      null,
      e,
      a
    ), t.push({ event: e, listeners: l }));
  }
  var cu = null, fu = null;
  function Nh(t) {
    Ko(t, 0);
  }
  function rn(t) {
    var l = lu(t);
    if (Kf(l)) return t;
  }
  function ds(t, l) {
    if (t === "change") return l;
  }
  var hs = !1;
  if (ae) {
    var Hi;
    if (ae) {
      var qi = "oninput" in document;
      if (!qi) {
        var ys = document.createElement("div");
        ys.setAttribute("oninput", "return;"), qi = typeof ys.oninput == "function";
      }
      Hi = qi;
    } else Hi = !1;
    hs = Hi && (!document.documentMode || 9 < document.documentMode);
  }
  function vs() {
    cu && (cu.detachEvent("onpropertychange", ms), fu = cu = null);
  }
  function ms(t) {
    if (t.propertyName === "value" && rn(fu)) {
      var l = [];
      os(
        l,
        fu,
        t,
        zi(t)
      ), If(Nh, l);
    }
  }
  function Dh(t, l, e) {
    t === "focusin" ? (vs(), cu = l, fu = e, cu.attachEvent("onpropertychange", ms)) : t === "focusout" && vs();
  }
  function Oh(t) {
    if (t === "selectionchange" || t === "keyup" || t === "keydown")
      return rn(fu);
  }
  function Rh(t, l) {
    if (t === "click") return rn(l);
  }
  function _h(t, l) {
    if (t === "input" || t === "change")
      return rn(l);
  }
  function Uh(t, l) {
    return t === l && (t !== 0 || 1 / t === 1 / l) || t !== t && l !== l;
  }
  var xl = typeof Object.is == "function" ? Object.is : Uh;
  function su(t, l) {
    if (xl(t, l)) return !0;
    if (typeof t != "object" || t === null || typeof l != "object" || l === null)
      return !1;
    var e = Object.keys(t), a = Object.keys(l);
    if (e.length !== a.length) return !1;
    for (a = 0; a < e.length; a++) {
      var u = e[a];
      if (!ml.call(l, u) || !xl(t[u], l[u]))
        return !1;
    }
    return !0;
  }
  function gs(t) {
    for (; t && t.firstChild; ) t = t.firstChild;
    return t;
  }
  function Ss(t, l) {
    var e = gs(t);
    t = 0;
    for (var a; e; ) {
      if (e.nodeType === 3) {
        if (a = t + e.textContent.length, t <= l && a >= l)
          return { node: e, offset: l - t };
        t = a;
      }
      t: {
        for (; e; ) {
          if (e.nextSibling) {
            e = e.nextSibling;
            break t;
          }
          e = e.parentNode;
        }
        e = void 0;
      }
      e = gs(e);
    }
  }
  function bs(t, l) {
    return t && l ? t === l ? !0 : t && t.nodeType === 3 ? !1 : l && l.nodeType === 3 ? bs(t, l.parentNode) : "contains" in t ? t.contains(l) : t.compareDocumentPosition ? !!(t.compareDocumentPosition(l) & 16) : !1 : !1;
  }
  function xs(t) {
    t = t != null && t.ownerDocument != null && t.ownerDocument.defaultView != null ? t.ownerDocument.defaultView : window;
    for (var l = en(t.document); l instanceof t.HTMLIFrameElement; ) {
      try {
        var e = typeof l.contentWindow.location.href == "string";
      } catch {
        e = !1;
      }
      if (e) t = l.contentWindow;
      else break;
      l = en(t.document);
    }
    return l;
  }
  function Bi(t) {
    var l = t && t.nodeName && t.nodeName.toLowerCase();
    return l && (l === "input" && (t.type === "text" || t.type === "search" || t.type === "tel" || t.type === "url" || t.type === "password") || l === "textarea" || t.contentEditable === "true");
  }
  var Ch = ae && "documentMode" in document && 11 >= document.documentMode, Aa = null, Yi = null, ru = null, Gi = !1;
  function ps(t, l, e) {
    var a = e.window === e ? e.document : e.nodeType === 9 ? e : e.ownerDocument;
    Gi || Aa == null || Aa !== en(a) || (a = Aa, "selectionStart" in a && Bi(a) ? a = { start: a.selectionStart, end: a.selectionEnd } : (a = (a.ownerDocument && a.ownerDocument.defaultView || window).getSelection(), a = {
      anchorNode: a.anchorNode,
      anchorOffset: a.anchorOffset,
      focusNode: a.focusNode,
      focusOffset: a.focusOffset
    }), ru && su(ru, a) || (ru = a, a = $n(Yi, "onSelect"), 0 < a.length && (l = new fn(
      "onSelect",
      "select",
      null,
      l,
      e
    ), t.push({ event: l, listeners: a }), l.target = Aa)));
  }
  function $e(t, l) {
    var e = {};
    return e[t.toLowerCase()] = l.toLowerCase(), e["Webkit" + t] = "webkit" + l, e["Moz" + t] = "moz" + l, e;
  }
  var Ea = {
    animationend: $e("Animation", "AnimationEnd"),
    animationiteration: $e("Animation", "AnimationIteration"),
    animationstart: $e("Animation", "AnimationStart"),
    transitionrun: $e("Transition", "TransitionRun"),
    transitionstart: $e("Transition", "TransitionStart"),
    transitioncancel: $e("Transition", "TransitionCancel"),
    transitionend: $e("Transition", "TransitionEnd")
  }, Xi = {}, js = {};
  ae && (js = document.createElement("div").style, "AnimationEvent" in window || (delete Ea.animationend.animation, delete Ea.animationiteration.animation, delete Ea.animationstart.animation), "TransitionEvent" in window || delete Ea.transitionend.transition);
  function We(t) {
    if (Xi[t]) return Xi[t];
    if (!Ea[t]) return t;
    var l = Ea[t], e;
    for (e in l)
      if (l.hasOwnProperty(e) && e in js)
        return Xi[t] = l[e];
    return t;
  }
  var Ts = We("animationend"), As = We("animationiteration"), Es = We("animationstart"), Hh = We("transitionrun"), qh = We("transitionstart"), Bh = We("transitioncancel"), zs = We("transitionend"), Ms = /* @__PURE__ */ new Map(), Qi = "abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(
    " "
  );
  Qi.push("scrollEnd");
  function Yl(t, l) {
    Ms.set(t, l), Je(l, [t]);
  }
  var Ns = /* @__PURE__ */ new WeakMap();
  function Ol(t, l) {
    if (typeof t == "object" && t !== null) {
      var e = Ns.get(t);
      return e !== void 0 ? e : (l = {
        value: t,
        source: l,
        stack: Vf(l)
      }, Ns.set(t, l), l);
    }
    return {
      value: t,
      source: l,
      stack: Vf(l)
    };
  }
  var Rl = [], za = 0, Zi = 0;
  function on() {
    for (var t = za, l = Zi = za = 0; l < t; ) {
      var e = Rl[l];
      Rl[l++] = null;
      var a = Rl[l];
      Rl[l++] = null;
      var u = Rl[l];
      Rl[l++] = null;
      var n = Rl[l];
      if (Rl[l++] = null, a !== null && u !== null) {
        var i = a.pending;
        i === null ? u.next = u : (u.next = i.next, i.next = u), a.pending = u;
      }
      n !== 0 && Ds(e, u, n);
    }
  }
  function dn(t, l, e, a) {
    Rl[za++] = t, Rl[za++] = l, Rl[za++] = e, Rl[za++] = a, Zi |= a, t.lanes |= a, t = t.alternate, t !== null && (t.lanes |= a);
  }
  function Li(t, l, e, a) {
    return dn(t, l, e, a), hn(t);
  }
  function Ma(t, l) {
    return dn(t, null, null, l), hn(t);
  }
  function Ds(t, l, e) {
    t.lanes |= e;
    var a = t.alternate;
    a !== null && (a.lanes |= e);
    for (var u = !1, n = t.return; n !== null; )
      n.childLanes |= e, a = n.alternate, a !== null && (a.childLanes |= e), n.tag === 22 && (t = n.stateNode, t === null || t._visibility & 1 || (u = !0)), t = n, n = n.return;
    return t.tag === 3 ? (n = t.stateNode, u && l !== null && (u = 31 - ct(e), t = n.hiddenUpdates, a = t[u], a === null ? t[u] = [l] : a.push(l), l.lane = e | 536870912), n) : null;
  }
  function hn(t) {
    if (50 < Hu)
      throw Hu = 0, $c = null, Error(o(185));
    for (var l = t.return; l !== null; )
      t = l, l = t.return;
    return t.tag === 3 ? t.stateNode : null;
  }
  var Na = {};
  function Yh(t, l, e, a) {
    this.tag = t, this.key = e, this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null, this.index = 0, this.refCleanup = this.ref = null, this.pendingProps = l, this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null, this.mode = a, this.subtreeFlags = this.flags = 0, this.deletions = null, this.childLanes = this.lanes = 0, this.alternate = null;
  }
  function pl(t, l, e, a) {
    return new Yh(t, l, e, a);
  }
  function Vi(t) {
    return t = t.prototype, !(!t || !t.isReactComponent);
  }
  function ue(t, l) {
    var e = t.alternate;
    return e === null ? (e = pl(
      t.tag,
      l,
      t.key,
      t.mode
    ), e.elementType = t.elementType, e.type = t.type, e.stateNode = t.stateNode, e.alternate = t, t.alternate = e) : (e.pendingProps = l, e.type = t.type, e.flags = 0, e.subtreeFlags = 0, e.deletions = null), e.flags = t.flags & 65011712, e.childLanes = t.childLanes, e.lanes = t.lanes, e.child = t.child, e.memoizedProps = t.memoizedProps, e.memoizedState = t.memoizedState, e.updateQueue = t.updateQueue, l = t.dependencies, e.dependencies = l === null ? null : { lanes: l.lanes, firstContext: l.firstContext }, e.sibling = t.sibling, e.index = t.index, e.ref = t.ref, e.refCleanup = t.refCleanup, e;
  }
  function Os(t, l) {
    t.flags &= 65011714;
    var e = t.alternate;
    return e === null ? (t.childLanes = 0, t.lanes = l, t.child = null, t.subtreeFlags = 0, t.memoizedProps = null, t.memoizedState = null, t.updateQueue = null, t.dependencies = null, t.stateNode = null) : (t.childLanes = e.childLanes, t.lanes = e.lanes, t.child = e.child, t.subtreeFlags = 0, t.deletions = null, t.memoizedProps = e.memoizedProps, t.memoizedState = e.memoizedState, t.updateQueue = e.updateQueue, t.type = e.type, l = e.dependencies, t.dependencies = l === null ? null : {
      lanes: l.lanes,
      firstContext: l.firstContext
    }), t;
  }
  function yn(t, l, e, a, u, n) {
    var i = 0;
    if (a = t, typeof t == "function") Vi(t) && (i = 1);
    else if (typeof t == "string")
      i = X0(
        t,
        e,
        L.current
      ) ? 26 : t === "html" || t === "head" || t === "body" ? 27 : 5;
    else
      t: switch (t) {
        case Wt:
          return t = pl(31, e, l, u), t.elementType = Wt, t.lanes = n, t;
        case B:
          return Fe(e.children, u, n, l);
        case rt:
          i = 8, u |= 24;
          break;
        case Ht:
          return t = pl(12, e, l, u | 2), t.elementType = Ht, t.lanes = n, t;
        case P:
          return t = pl(13, e, l, u), t.elementType = P, t.lanes = n, t;
        case $t:
          return t = pl(19, e, l, u), t.elementType = $t, t.lanes = n, t;
        default:
          if (typeof t == "object" && t !== null)
            switch (t.$$typeof) {
              case xt:
              case Et:
                i = 10;
                break t;
              case qt:
                i = 9;
                break t;
              case nt:
                i = 11;
                break t;
              case V:
                i = 14;
                break t;
              case zt:
                i = 16, a = null;
                break t;
            }
          i = 29, e = Error(
            o(130, t === null ? "null" : typeof t, "")
          ), a = null;
      }
    return l = pl(i, e, l, u), l.elementType = t, l.type = a, l.lanes = n, l;
  }
  function Fe(t, l, e, a) {
    return t = pl(7, t, a, l), t.lanes = e, t;
  }
  function wi(t, l, e) {
    return t = pl(6, t, null, l), t.lanes = e, t;
  }
  function Ki(t, l, e) {
    return l = pl(
      4,
      t.children !== null ? t.children : [],
      t.key,
      l
    ), l.lanes = e, l.stateNode = {
      containerInfo: t.containerInfo,
      pendingChildren: null,
      implementation: t.implementation
    }, l;
  }
  var Da = [], Oa = 0, vn = null, mn = 0, _l = [], Ul = 0, Pe = null, ne = 1, ie = "";
  function Ie(t, l) {
    Da[Oa++] = mn, Da[Oa++] = vn, vn = t, mn = l;
  }
  function Rs(t, l, e) {
    _l[Ul++] = ne, _l[Ul++] = ie, _l[Ul++] = Pe, Pe = t;
    var a = ne;
    t = ie;
    var u = 32 - ct(a) - 1;
    a &= ~(1 << u), e += 1;
    var n = 32 - ct(l) + u;
    if (30 < n) {
      var i = u - u % 5;
      n = (a & (1 << i) - 1).toString(32), a >>= i, u -= i, ne = 1 << 32 - ct(l) + u | e << u | a, ie = n + t;
    } else
      ne = 1 << n | e << u | a, ie = t;
  }
  function Ji(t) {
    t.return !== null && (Ie(t, 1), Rs(t, 1, 0));
  }
  function ki(t) {
    for (; t === vn; )
      vn = Da[--Oa], Da[Oa] = null, mn = Da[--Oa], Da[Oa] = null;
    for (; t === Pe; )
      Pe = _l[--Ul], _l[Ul] = null, ie = _l[--Ul], _l[Ul] = null, ne = _l[--Ul], _l[Ul] = null;
  }
  var fl = null, Ot = null, yt = !1, ta = null, Kl = !1, $i = Error(o(519));
  function la(t) {
    var l = Error(o(418, ""));
    throw hu(Ol(l, t)), $i;
  }
  function _s(t) {
    var l = t.stateNode, e = t.type, a = t.memoizedProps;
    switch (l[el] = t, l[sl] = a, e) {
      case "dialog":
        ut("cancel", l), ut("close", l);
        break;
      case "iframe":
      case "object":
      case "embed":
        ut("load", l);
        break;
      case "video":
      case "audio":
        for (e = 0; e < Bu.length; e++)
          ut(Bu[e], l);
        break;
      case "source":
        ut("error", l);
        break;
      case "img":
      case "image":
      case "link":
        ut("error", l), ut("load", l);
        break;
      case "details":
        ut("toggle", l);
        break;
      case "input":
        ut("invalid", l), Jf(
          l,
          a.value,
          a.defaultValue,
          a.checked,
          a.defaultChecked,
          a.type,
          a.name,
          !0
        ), ln(l);
        break;
      case "select":
        ut("invalid", l);
        break;
      case "textarea":
        ut("invalid", l), $f(l, a.value, a.defaultValue, a.children), ln(l);
    }
    e = a.children, typeof e != "string" && typeof e != "number" && typeof e != "bigint" || l.textContent === "" + e || a.suppressHydrationWarning === !0 || Wo(l.textContent, e) ? (a.popover != null && (ut("beforetoggle", l), ut("toggle", l)), a.onScroll != null && ut("scroll", l), a.onScrollEnd != null && ut("scrollend", l), a.onClick != null && (l.onclick = Wn), l = !0) : l = !1, l || la(t);
  }
  function Us(t) {
    for (fl = t.return; fl; )
      switch (fl.tag) {
        case 5:
        case 13:
          Kl = !1;
          return;
        case 27:
        case 3:
          Kl = !0;
          return;
        default:
          fl = fl.return;
      }
  }
  function ou(t) {
    if (t !== fl) return !1;
    if (!yt) return Us(t), yt = !0, !1;
    var l = t.tag, e;
    if ((e = l !== 3 && l !== 27) && ((e = l === 5) && (e = t.type, e = !(e !== "form" && e !== "button") || df(t.type, t.memoizedProps)), e = !e), e && Ot && la(t), Us(t), l === 13) {
      if (t = t.memoizedState, t = t !== null ? t.dehydrated : null, !t) throw Error(o(317));
      t: {
        for (t = t.nextSibling, l = 0; t; ) {
          if (t.nodeType === 8)
            if (e = t.data, e === "/$") {
              if (l === 0) {
                Ot = Xl(t.nextSibling);
                break t;
              }
              l--;
            } else
              e !== "$" && e !== "$!" && e !== "$?" || l++;
          t = t.nextSibling;
        }
        Ot = null;
      }
    } else
      l === 27 ? (l = Ot, qe(t.type) ? (t = mf, mf = null, Ot = t) : Ot = l) : Ot = fl ? Xl(t.stateNode.nextSibling) : null;
    return !0;
  }
  function du() {
    Ot = fl = null, yt = !1;
  }
  function Cs() {
    var t = ta;
    return t !== null && (hl === null ? hl = t : hl.push.apply(
      hl,
      t
    ), ta = null), t;
  }
  function hu(t) {
    ta === null ? ta = [t] : ta.push(t);
  }
  var Wi = M(null), ea = null, ce = null;
  function pe(t, l, e) {
    q(Wi, l._currentValue), l._currentValue = e;
  }
  function fe(t) {
    t._currentValue = Wi.current, Y(Wi);
  }
  function Fi(t, l, e) {
    for (; t !== null; ) {
      var a = t.alternate;
      if ((t.childLanes & l) !== l ? (t.childLanes |= l, a !== null && (a.childLanes |= l)) : a !== null && (a.childLanes & l) !== l && (a.childLanes |= l), t === e) break;
      t = t.return;
    }
  }
  function Pi(t, l, e, a) {
    var u = t.child;
    for (u !== null && (u.return = t); u !== null; ) {
      var n = u.dependencies;
      if (n !== null) {
        var i = u.child;
        n = n.firstContext;
        t: for (; n !== null; ) {
          var f = n;
          n = u;
          for (var s = 0; s < l.length; s++)
            if (f.context === l[s]) {
              n.lanes |= e, f = n.alternate, f !== null && (f.lanes |= e), Fi(
                n.return,
                e,
                t
              ), a || (i = null);
              break t;
            }
          n = f.next;
        }
      } else if (u.tag === 18) {
        if (i = u.return, i === null) throw Error(o(341));
        i.lanes |= e, n = i.alternate, n !== null && (n.lanes |= e), Fi(i, e, t), i = null;
      } else i = u.child;
      if (i !== null) i.return = u;
      else
        for (i = u; i !== null; ) {
          if (i === t) {
            i = null;
            break;
          }
          if (u = i.sibling, u !== null) {
            u.return = i.return, i = u;
            break;
          }
          i = i.return;
        }
      u = i;
    }
  }
  function yu(t, l, e, a) {
    t = null;
    for (var u = l, n = !1; u !== null; ) {
      if (!n) {
        if ((u.flags & 524288) !== 0) n = !0;
        else if ((u.flags & 262144) !== 0) break;
      }
      if (u.tag === 10) {
        var i = u.alternate;
        if (i === null) throw Error(o(387));
        if (i = i.memoizedProps, i !== null) {
          var f = u.type;
          xl(u.pendingProps.value, i.value) || (t !== null ? t.push(f) : t = [f]);
        }
      } else if (u === ll.current) {
        if (i = u.alternate, i === null) throw Error(o(387));
        i.memoizedState.memoizedState !== u.memoizedState.memoizedState && (t !== null ? t.push(Lu) : t = [Lu]);
      }
      u = u.return;
    }
    t !== null && Pi(
      l,
      t,
      e,
      a
    ), l.flags |= 262144;
  }
  function gn(t) {
    for (t = t.firstContext; t !== null; ) {
      if (!xl(
        t.context._currentValue,
        t.memoizedValue
      ))
        return !0;
      t = t.next;
    }
    return !1;
  }
  function aa(t) {
    ea = t, ce = null, t = t.dependencies, t !== null && (t.firstContext = null);
  }
  function al(t) {
    return Hs(ea, t);
  }
  function Sn(t, l) {
    return ea === null && aa(t), Hs(t, l);
  }
  function Hs(t, l) {
    var e = l._currentValue;
    if (l = { context: l, memoizedValue: e, next: null }, ce === null) {
      if (t === null) throw Error(o(308));
      ce = l, t.dependencies = { lanes: 0, firstContext: l }, t.flags |= 524288;
    } else ce = ce.next = l;
    return e;
  }
  var Gh = typeof AbortController < "u" ? AbortController : function() {
    var t = [], l = this.signal = {
      aborted: !1,
      addEventListener: function(e, a) {
        t.push(a);
      }
    };
    this.abort = function() {
      l.aborted = !0, t.forEach(function(e) {
        return e();
      });
    };
  }, Xh = r.unstable_scheduleCallback, Qh = r.unstable_NormalPriority, Gt = {
    $$typeof: Et,
    Consumer: null,
    Provider: null,
    _currentValue: null,
    _currentValue2: null,
    _threadCount: 0
  };
  function Ii() {
    return {
      controller: new Gh(),
      data: /* @__PURE__ */ new Map(),
      refCount: 0
    };
  }
  function vu(t) {
    t.refCount--, t.refCount === 0 && Xh(Qh, function() {
      t.controller.abort();
    });
  }
  var mu = null, tc = 0, Ra = 0, _a = null;
  function Zh(t, l) {
    if (mu === null) {
      var e = mu = [];
      tc = 0, Ra = ef(), _a = {
        status: "pending",
        value: void 0,
        then: function(a) {
          e.push(a);
        }
      };
    }
    return tc++, l.then(qs, qs), l;
  }
  function qs() {
    if (--tc === 0 && mu !== null) {
      _a !== null && (_a.status = "fulfilled");
      var t = mu;
      mu = null, Ra = 0, _a = null;
      for (var l = 0; l < t.length; l++) (0, t[l])();
    }
  }
  function Lh(t, l) {
    var e = [], a = {
      status: "pending",
      value: null,
      reason: null,
      then: function(u) {
        e.push(u);
      }
    };
    return t.then(
      function() {
        a.status = "fulfilled", a.value = l;
        for (var u = 0; u < e.length; u++) (0, e[u])(l);
      },
      function(u) {
        for (a.status = "rejected", a.reason = u, u = 0; u < e.length; u++)
          (0, e[u])(void 0);
      }
    ), a;
  }
  var Bs = p.S;
  p.S = function(t, l) {
    typeof l == "object" && l !== null && typeof l.then == "function" && Zh(t, l), Bs !== null && Bs(t, l);
  };
  var ua = M(null);
  function lc() {
    var t = ua.current;
    return t !== null ? t : At.pooledCache;
  }
  function bn(t, l) {
    l === null ? q(ua, ua.current) : q(ua, l.pool);
  }
  function Ys() {
    var t = lc();
    return t === null ? null : { parent: Gt._currentValue, pool: t };
  }
  var gu = Error(o(460)), Gs = Error(o(474)), xn = Error(o(542)), ec = { then: function() {
  } };
  function Xs(t) {
    return t = t.status, t === "fulfilled" || t === "rejected";
  }
  function pn() {
  }
  function Qs(t, l, e) {
    switch (e = t[e], e === void 0 ? t.push(l) : e !== l && (l.then(pn, pn), l = e), l.status) {
      case "fulfilled":
        return l.value;
      case "rejected":
        throw t = l.reason, Ls(t), t;
      default:
        if (typeof l.status == "string") l.then(pn, pn);
        else {
          if (t = At, t !== null && 100 < t.shellSuspendCounter)
            throw Error(o(482));
          t = l, t.status = "pending", t.then(
            function(a) {
              if (l.status === "pending") {
                var u = l;
                u.status = "fulfilled", u.value = a;
              }
            },
            function(a) {
              if (l.status === "pending") {
                var u = l;
                u.status = "rejected", u.reason = a;
              }
            }
          );
        }
        switch (l.status) {
          case "fulfilled":
            return l.value;
          case "rejected":
            throw t = l.reason, Ls(t), t;
        }
        throw Su = l, gu;
    }
  }
  var Su = null;
  function Zs() {
    if (Su === null) throw Error(o(459));
    var t = Su;
    return Su = null, t;
  }
  function Ls(t) {
    if (t === gu || t === xn)
      throw Error(o(483));
  }
  var je = !1;
  function ac(t) {
    t.updateQueue = {
      baseState: t.memoizedState,
      firstBaseUpdate: null,
      lastBaseUpdate: null,
      shared: { pending: null, lanes: 0, hiddenCallbacks: null },
      callbacks: null
    };
  }
  function uc(t, l) {
    t = t.updateQueue, l.updateQueue === t && (l.updateQueue = {
      baseState: t.baseState,
      firstBaseUpdate: t.firstBaseUpdate,
      lastBaseUpdate: t.lastBaseUpdate,
      shared: t.shared,
      callbacks: null
    });
  }
  function Te(t) {
    return { lane: t, tag: 0, payload: null, callback: null, next: null };
  }
  function Ae(t, l, e) {
    var a = t.updateQueue;
    if (a === null) return null;
    if (a = a.shared, (vt & 2) !== 0) {
      var u = a.pending;
      return u === null ? l.next = l : (l.next = u.next, u.next = l), a.pending = l, l = hn(t), Ds(t, null, e), l;
    }
    return dn(t, a, l, e), hn(t);
  }
  function bu(t, l, e) {
    if (l = l.updateQueue, l !== null && (l = l.shared, (e & 4194048) !== 0)) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, Ke(t, e);
    }
  }
  function nc(t, l) {
    var e = t.updateQueue, a = t.alternate;
    if (a !== null && (a = a.updateQueue, e === a)) {
      var u = null, n = null;
      if (e = e.firstBaseUpdate, e !== null) {
        do {
          var i = {
            lane: e.lane,
            tag: e.tag,
            payload: e.payload,
            callback: null,
            next: null
          };
          n === null ? u = n = i : n = n.next = i, e = e.next;
        } while (e !== null);
        n === null ? u = n = l : n = n.next = l;
      } else u = n = l;
      e = {
        baseState: a.baseState,
        firstBaseUpdate: u,
        lastBaseUpdate: n,
        shared: a.shared,
        callbacks: a.callbacks
      }, t.updateQueue = e;
      return;
    }
    t = e.lastBaseUpdate, t === null ? e.firstBaseUpdate = l : t.next = l, e.lastBaseUpdate = l;
  }
  var ic = !1;
  function xu() {
    if (ic) {
      var t = _a;
      if (t !== null) throw t;
    }
  }
  function pu(t, l, e, a) {
    ic = !1;
    var u = t.updateQueue;
    je = !1;
    var n = u.firstBaseUpdate, i = u.lastBaseUpdate, f = u.shared.pending;
    if (f !== null) {
      u.shared.pending = null;
      var s = f, m = s.next;
      s.next = null, i === null ? n = m : i.next = m, i = s;
      var j = t.alternate;
      j !== null && (j = j.updateQueue, f = j.lastBaseUpdate, f !== i && (f === null ? j.firstBaseUpdate = m : f.next = m, j.lastBaseUpdate = s));
    }
    if (n !== null) {
      var z = u.baseState;
      i = 0, j = m = s = null, f = n;
      do {
        var g = f.lane & -536870913, S = g !== f.lane;
        if (S ? (ft & g) === g : (a & g) === g) {
          g !== 0 && g === Ra && (ic = !0), j !== null && (j = j.next = {
            lane: 0,
            tag: f.tag,
            payload: f.payload,
            callback: null,
            next: null
          });
          t: {
            var k = t, K = f;
            g = l;
            var bt = e;
            switch (K.tag) {
              case 1:
                if (k = K.payload, typeof k == "function") {
                  z = k.call(bt, z, g);
                  break t;
                }
                z = k;
                break t;
              case 3:
                k.flags = k.flags & -65537 | 128;
              case 0:
                if (k = K.payload, g = typeof k == "function" ? k.call(bt, z, g) : k, g == null) break t;
                z = C({}, z, g);
                break t;
              case 2:
                je = !0;
            }
          }
          g = f.callback, g !== null && (t.flags |= 64, S && (t.flags |= 8192), S = u.callbacks, S === null ? u.callbacks = [g] : S.push(g));
        } else
          S = {
            lane: g,
            tag: f.tag,
            payload: f.payload,
            callback: f.callback,
            next: null
          }, j === null ? (m = j = S, s = z) : j = j.next = S, i |= g;
        if (f = f.next, f === null) {
          if (f = u.shared.pending, f === null)
            break;
          S = f, f = S.next, S.next = null, u.lastBaseUpdate = S, u.shared.pending = null;
        }
      } while (!0);
      j === null && (s = z), u.baseState = s, u.firstBaseUpdate = m, u.lastBaseUpdate = j, n === null && (u.shared.lanes = 0), _e |= i, t.lanes = i, t.memoizedState = z;
    }
  }
  function Vs(t, l) {
    if (typeof t != "function")
      throw Error(o(191, t));
    t.call(l);
  }
  function ws(t, l) {
    var e = t.callbacks;
    if (e !== null)
      for (t.callbacks = null, t = 0; t < e.length; t++)
        Vs(e[t], l);
  }
  var Ua = M(null), jn = M(0);
  function Ks(t, l) {
    t = ve, q(jn, t), q(Ua, l), ve = t | l.baseLanes;
  }
  function cc() {
    q(jn, ve), q(Ua, Ua.current);
  }
  function fc() {
    ve = jn.current, Y(Ua), Y(jn);
  }
  var Ee = 0, lt = null, gt = null, Bt = null, Tn = !1, Ca = !1, na = !1, An = 0, ju = 0, Ha = null, Vh = 0;
  function Ut() {
    throw Error(o(321));
  }
  function sc(t, l) {
    if (l === null) return !1;
    for (var e = 0; e < l.length && e < t.length; e++)
      if (!xl(t[e], l[e])) return !1;
    return !0;
  }
  function rc(t, l, e, a, u, n) {
    return Ee = n, lt = l, l.memoizedState = null, l.updateQueue = null, l.lanes = 0, p.H = t === null || t.memoizedState === null ? Dr : Or, na = !1, n = e(a, u), na = !1, Ca && (n = ks(
      l,
      e,
      a,
      u
    )), Js(t), n;
  }
  function Js(t) {
    p.H = On;
    var l = gt !== null && gt.next !== null;
    if (Ee = 0, Bt = gt = lt = null, Tn = !1, ju = 0, Ha = null, l) throw Error(o(300));
    t === null || Kt || (t = t.dependencies, t !== null && gn(t) && (Kt = !0));
  }
  function ks(t, l, e, a) {
    lt = t;
    var u = 0;
    do {
      if (Ca && (Ha = null), ju = 0, Ca = !1, 25 <= u) throw Error(o(301));
      if (u += 1, Bt = gt = null, t.updateQueue != null) {
        var n = t.updateQueue;
        n.lastEffect = null, n.events = null, n.stores = null, n.memoCache != null && (n.memoCache.index = 0);
      }
      p.H = Fh, n = l(e, a);
    } while (Ca);
    return n;
  }
  function wh() {
    var t = p.H, l = t.useState()[0];
    return l = typeof l.then == "function" ? Tu(l) : l, t = t.useState()[0], (gt !== null ? gt.memoizedState : null) !== t && (lt.flags |= 1024), l;
  }
  function oc() {
    var t = An !== 0;
    return An = 0, t;
  }
  function dc(t, l, e) {
    l.updateQueue = t.updateQueue, l.flags &= -2053, t.lanes &= ~e;
  }
  function hc(t) {
    if (Tn) {
      for (t = t.memoizedState; t !== null; ) {
        var l = t.queue;
        l !== null && (l.pending = null), t = t.next;
      }
      Tn = !1;
    }
    Ee = 0, Bt = gt = lt = null, Ca = !1, ju = An = 0, Ha = null;
  }
  function ol() {
    var t = {
      memoizedState: null,
      baseState: null,
      baseQueue: null,
      queue: null,
      next: null
    };
    return Bt === null ? lt.memoizedState = Bt = t : Bt = Bt.next = t, Bt;
  }
  function Yt() {
    if (gt === null) {
      var t = lt.alternate;
      t = t !== null ? t.memoizedState : null;
    } else t = gt.next;
    var l = Bt === null ? lt.memoizedState : Bt.next;
    if (l !== null)
      Bt = l, gt = t;
    else {
      if (t === null)
        throw lt.alternate === null ? Error(o(467)) : Error(o(310));
      gt = t, t = {
        memoizedState: gt.memoizedState,
        baseState: gt.baseState,
        baseQueue: gt.baseQueue,
        queue: gt.queue,
        next: null
      }, Bt === null ? lt.memoizedState = Bt = t : Bt = Bt.next = t;
    }
    return Bt;
  }
  function yc() {
    return { lastEffect: null, events: null, stores: null, memoCache: null };
  }
  function Tu(t) {
    var l = ju;
    return ju += 1, Ha === null && (Ha = []), t = Qs(Ha, t, l), l = lt, (Bt === null ? l.memoizedState : Bt.next) === null && (l = l.alternate, p.H = l === null || l.memoizedState === null ? Dr : Or), t;
  }
  function En(t) {
    if (t !== null && typeof t == "object") {
      if (typeof t.then == "function") return Tu(t);
      if (t.$$typeof === Et) return al(t);
    }
    throw Error(o(438, String(t)));
  }
  function vc(t) {
    var l = null, e = lt.updateQueue;
    if (e !== null && (l = e.memoCache), l == null) {
      var a = lt.alternate;
      a !== null && (a = a.updateQueue, a !== null && (a = a.memoCache, a != null && (l = {
        data: a.data.map(function(u) {
          return u.slice();
        }),
        index: 0
      })));
    }
    if (l == null && (l = { data: [], index: 0 }), e === null && (e = yc(), lt.updateQueue = e), e.memoCache = l, e = l.data[l.index], e === void 0)
      for (e = l.data[l.index] = Array(t), a = 0; a < t; a++)
        e[a] = Fl;
    return l.index++, e;
  }
  function se(t, l) {
    return typeof l == "function" ? l(t) : l;
  }
  function zn(t) {
    var l = Yt();
    return mc(l, gt, t);
  }
  function mc(t, l, e) {
    var a = t.queue;
    if (a === null) throw Error(o(311));
    a.lastRenderedReducer = e;
    var u = t.baseQueue, n = a.pending;
    if (n !== null) {
      if (u !== null) {
        var i = u.next;
        u.next = n.next, n.next = i;
      }
      l.baseQueue = u = n, a.pending = null;
    }
    if (n = t.baseState, u === null) t.memoizedState = n;
    else {
      l = u.next;
      var f = i = null, s = null, m = l, j = !1;
      do {
        var z = m.lane & -536870913;
        if (z !== m.lane ? (ft & z) === z : (Ee & z) === z) {
          var g = m.revertLane;
          if (g === 0)
            s !== null && (s = s.next = {
              lane: 0,
              revertLane: 0,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }), z === Ra && (j = !0);
          else if ((Ee & g) === g) {
            m = m.next, g === Ra && (j = !0);
            continue;
          } else
            z = {
              lane: 0,
              revertLane: m.revertLane,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }, s === null ? (f = s = z, i = n) : s = s.next = z, lt.lanes |= g, _e |= g;
          z = m.action, na && e(n, z), n = m.hasEagerState ? m.eagerState : e(n, z);
        } else
          g = {
            lane: z,
            revertLane: m.revertLane,
            action: m.action,
            hasEagerState: m.hasEagerState,
            eagerState: m.eagerState,
            next: null
          }, s === null ? (f = s = g, i = n) : s = s.next = g, lt.lanes |= z, _e |= z;
        m = m.next;
      } while (m !== null && m !== l);
      if (s === null ? i = n : s.next = f, !xl(n, t.memoizedState) && (Kt = !0, j && (e = _a, e !== null)))
        throw e;
      t.memoizedState = n, t.baseState = i, t.baseQueue = s, a.lastRenderedState = n;
    }
    return u === null && (a.lanes = 0), [t.memoizedState, a.dispatch];
  }
  function gc(t) {
    var l = Yt(), e = l.queue;
    if (e === null) throw Error(o(311));
    e.lastRenderedReducer = t;
    var a = e.dispatch, u = e.pending, n = l.memoizedState;
    if (u !== null) {
      e.pending = null;
      var i = u = u.next;
      do
        n = t(n, i.action), i = i.next;
      while (i !== u);
      xl(n, l.memoizedState) || (Kt = !0), l.memoizedState = n, l.baseQueue === null && (l.baseState = n), e.lastRenderedState = n;
    }
    return [n, a];
  }
  function $s(t, l, e) {
    var a = lt, u = Yt(), n = yt;
    if (n) {
      if (e === void 0) throw Error(o(407));
      e = e();
    } else e = l();
    var i = !xl(
      (gt || u).memoizedState,
      e
    );
    i && (u.memoizedState = e, Kt = !0), u = u.queue;
    var f = Ps.bind(null, a, u, t);
    if (Au(2048, 8, f, [t]), u.getSnapshot !== l || i || Bt !== null && Bt.memoizedState.tag & 1) {
      if (a.flags |= 2048, qa(
        9,
        Mn(),
        Fs.bind(
          null,
          a,
          u,
          e,
          l
        ),
        null
      ), At === null) throw Error(o(349));
      n || (Ee & 124) !== 0 || Ws(a, l, e);
    }
    return e;
  }
  function Ws(t, l, e) {
    t.flags |= 16384, t = { getSnapshot: l, value: e }, l = lt.updateQueue, l === null ? (l = yc(), lt.updateQueue = l, l.stores = [t]) : (e = l.stores, e === null ? l.stores = [t] : e.push(t));
  }
  function Fs(t, l, e, a) {
    l.value = e, l.getSnapshot = a, Is(l) && tr(t);
  }
  function Ps(t, l, e) {
    return e(function() {
      Is(l) && tr(t);
    });
  }
  function Is(t) {
    var l = t.getSnapshot;
    t = t.value;
    try {
      var e = l();
      return !xl(t, e);
    } catch {
      return !0;
    }
  }
  function tr(t) {
    var l = Ma(t, 2);
    l !== null && zl(l, t, 2);
  }
  function Sc(t) {
    var l = ol();
    if (typeof t == "function") {
      var e = t;
      if (t = e(), na) {
        il(!0);
        try {
          e();
        } finally {
          il(!1);
        }
      }
    }
    return l.memoizedState = l.baseState = t, l.queue = {
      pending: null,
      lanes: 0,
      dispatch: null,
      lastRenderedReducer: se,
      lastRenderedState: t
    }, l;
  }
  function lr(t, l, e, a) {
    return t.baseState = e, mc(
      t,
      gt,
      typeof a == "function" ? a : se
    );
  }
  function Kh(t, l, e, a, u) {
    if (Dn(t)) throw Error(o(485));
    if (t = l.action, t !== null) {
      var n = {
        payload: u,
        action: t,
        next: null,
        isTransition: !0,
        status: "pending",
        value: null,
        reason: null,
        listeners: [],
        then: function(i) {
          n.listeners.push(i);
        }
      };
      p.T !== null ? e(!0) : n.isTransition = !1, a(n), e = l.pending, e === null ? (n.next = l.pending = n, er(l, n)) : (n.next = e.next, l.pending = e.next = n);
    }
  }
  function er(t, l) {
    var e = l.action, a = l.payload, u = t.state;
    if (l.isTransition) {
      var n = p.T, i = {};
      p.T = i;
      try {
        var f = e(u, a), s = p.S;
        s !== null && s(i, f), ar(t, l, f);
      } catch (m) {
        bc(t, l, m);
      } finally {
        p.T = n;
      }
    } else
      try {
        n = e(u, a), ar(t, l, n);
      } catch (m) {
        bc(t, l, m);
      }
  }
  function ar(t, l, e) {
    e !== null && typeof e == "object" && typeof e.then == "function" ? e.then(
      function(a) {
        ur(t, l, a);
      },
      function(a) {
        return bc(t, l, a);
      }
    ) : ur(t, l, e);
  }
  function ur(t, l, e) {
    l.status = "fulfilled", l.value = e, nr(l), t.state = e, l = t.pending, l !== null && (e = l.next, e === l ? t.pending = null : (e = e.next, l.next = e, er(t, e)));
  }
  function bc(t, l, e) {
    var a = t.pending;
    if (t.pending = null, a !== null) {
      a = a.next;
      do
        l.status = "rejected", l.reason = e, nr(l), l = l.next;
      while (l !== a);
    }
    t.action = null;
  }
  function nr(t) {
    t = t.listeners;
    for (var l = 0; l < t.length; l++) (0, t[l])();
  }
  function ir(t, l) {
    return l;
  }
  function cr(t, l) {
    if (yt) {
      var e = At.formState;
      if (e !== null) {
        t: {
          var a = lt;
          if (yt) {
            if (Ot) {
              l: {
                for (var u = Ot, n = Kl; u.nodeType !== 8; ) {
                  if (!n) {
                    u = null;
                    break l;
                  }
                  if (u = Xl(
                    u.nextSibling
                  ), u === null) {
                    u = null;
                    break l;
                  }
                }
                n = u.data, u = n === "F!" || n === "F" ? u : null;
              }
              if (u) {
                Ot = Xl(
                  u.nextSibling
                ), a = u.data === "F!";
                break t;
              }
            }
            la(a);
          }
          a = !1;
        }
        a && (l = e[0]);
      }
    }
    return e = ol(), e.memoizedState = e.baseState = l, a = {
      pending: null,
      lanes: 0,
      dispatch: null,
      lastRenderedReducer: ir,
      lastRenderedState: l
    }, e.queue = a, e = zr.bind(
      null,
      lt,
      a
    ), a.dispatch = e, a = Sc(!1), n = Ac.bind(
      null,
      lt,
      !1,
      a.queue
    ), a = ol(), u = {
      state: l,
      dispatch: null,
      action: t,
      pending: null
    }, a.queue = u, e = Kh.bind(
      null,
      lt,
      u,
      n,
      e
    ), u.dispatch = e, a.memoizedState = t, [l, e, !1];
  }
  function fr(t) {
    var l = Yt();
    return sr(l, gt, t);
  }
  function sr(t, l, e) {
    if (l = mc(
      t,
      l,
      ir
    )[0], t = zn(se)[0], typeof l == "object" && l !== null && typeof l.then == "function")
      try {
        var a = Tu(l);
      } catch (i) {
        throw i === gu ? xn : i;
      }
    else a = l;
    l = Yt();
    var u = l.queue, n = u.dispatch;
    return e !== l.memoizedState && (lt.flags |= 2048, qa(
      9,
      Mn(),
      Jh.bind(null, u, e),
      null
    )), [a, n, t];
  }
  function Jh(t, l) {
    t.action = l;
  }
  function rr(t) {
    var l = Yt(), e = gt;
    if (e !== null)
      return sr(l, e, t);
    Yt(), l = l.memoizedState, e = Yt();
    var a = e.queue.dispatch;
    return e.memoizedState = t, [l, a, !1];
  }
  function qa(t, l, e, a) {
    return t = { tag: t, create: e, deps: a, inst: l, next: null }, l = lt.updateQueue, l === null && (l = yc(), lt.updateQueue = l), e = l.lastEffect, e === null ? l.lastEffect = t.next = t : (a = e.next, e.next = t, t.next = a, l.lastEffect = t), t;
  }
  function Mn() {
    return { destroy: void 0, resource: void 0 };
  }
  function or() {
    return Yt().memoizedState;
  }
  function Nn(t, l, e, a) {
    var u = ol();
    a = a === void 0 ? null : a, lt.flags |= t, u.memoizedState = qa(
      1 | l,
      Mn(),
      e,
      a
    );
  }
  function Au(t, l, e, a) {
    var u = Yt();
    a = a === void 0 ? null : a;
    var n = u.memoizedState.inst;
    gt !== null && a !== null && sc(a, gt.memoizedState.deps) ? u.memoizedState = qa(l, n, e, a) : (lt.flags |= t, u.memoizedState = qa(
      1 | l,
      n,
      e,
      a
    ));
  }
  function dr(t, l) {
    Nn(8390656, 8, t, l);
  }
  function hr(t, l) {
    Au(2048, 8, t, l);
  }
  function yr(t, l) {
    return Au(4, 2, t, l);
  }
  function vr(t, l) {
    return Au(4, 4, t, l);
  }
  function mr(t, l) {
    if (typeof l == "function") {
      t = t();
      var e = l(t);
      return function() {
        typeof e == "function" ? e() : l(null);
      };
    }
    if (l != null)
      return t = t(), l.current = t, function() {
        l.current = null;
      };
  }
  function gr(t, l, e) {
    e = e != null ? e.concat([t]) : null, Au(4, 4, mr.bind(null, l, t), e);
  }
  function xc() {
  }
  function Sr(t, l) {
    var e = Yt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    return l !== null && sc(l, a[1]) ? a[0] : (e.memoizedState = [t, l], t);
  }
  function br(t, l) {
    var e = Yt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    if (l !== null && sc(l, a[1]))
      return a[0];
    if (a = t(), na) {
      il(!0);
      try {
        t();
      } finally {
        il(!1);
      }
    }
    return e.memoizedState = [a, l], a;
  }
  function pc(t, l, e) {
    return e === void 0 || (Ee & 1073741824) !== 0 ? t.memoizedState = l : (t.memoizedState = e, t = To(), lt.lanes |= t, _e |= t, e);
  }
  function xr(t, l, e, a) {
    return xl(e, l) ? e : Ua.current !== null ? (t = pc(t, e, a), xl(t, l) || (Kt = !0), t) : (Ee & 42) === 0 ? (Kt = !0, t.memoizedState = e) : (t = To(), lt.lanes |= t, _e |= t, l);
  }
  function pr(t, l, e, a, u) {
    var n = H.p;
    H.p = n !== 0 && 8 > n ? n : 8;
    var i = p.T, f = {};
    p.T = f, Ac(t, !1, l, e);
    try {
      var s = u(), m = p.S;
      if (m !== null && m(f, s), s !== null && typeof s == "object" && typeof s.then == "function") {
        var j = Lh(
          s,
          a
        );
        Eu(
          t,
          l,
          j,
          El(t)
        );
      } else
        Eu(
          t,
          l,
          a,
          El(t)
        );
    } catch (z) {
      Eu(
        t,
        l,
        { then: function() {
        }, status: "rejected", reason: z },
        El()
      );
    } finally {
      H.p = n, p.T = i;
    }
  }
  function kh() {
  }
  function jc(t, l, e, a) {
    if (t.tag !== 5) throw Error(o(476));
    var u = jr(t).queue;
    pr(
      t,
      u,
      l,
      w,
      e === null ? kh : function() {
        return Tr(t), e(a);
      }
    );
  }
  function jr(t) {
    var l = t.memoizedState;
    if (l !== null) return l;
    l = {
      memoizedState: w,
      baseState: w,
      baseQueue: null,
      queue: {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: se,
        lastRenderedState: w
      },
      next: null
    };
    var e = {};
    return l.next = {
      memoizedState: e,
      baseState: e,
      baseQueue: null,
      queue: {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: se,
        lastRenderedState: e
      },
      next: null
    }, t.memoizedState = l, t = t.alternate, t !== null && (t.memoizedState = l), l;
  }
  function Tr(t) {
    var l = jr(t).next.queue;
    Eu(t, l, {}, El());
  }
  function Tc() {
    return al(Lu);
  }
  function Ar() {
    return Yt().memoizedState;
  }
  function Er() {
    return Yt().memoizedState;
  }
  function $h(t) {
    for (var l = t.return; l !== null; ) {
      switch (l.tag) {
        case 24:
        case 3:
          var e = El();
          t = Te(e);
          var a = Ae(l, t, e);
          a !== null && (zl(a, l, e), bu(a, l, e)), l = { cache: Ii() }, t.payload = l;
          return;
      }
      l = l.return;
    }
  }
  function Wh(t, l, e) {
    var a = El();
    e = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, Dn(t) ? Mr(l, e) : (e = Li(t, l, e, a), e !== null && (zl(e, t, a), Nr(e, l, a)));
  }
  function zr(t, l, e) {
    var a = El();
    Eu(t, l, e, a);
  }
  function Eu(t, l, e, a) {
    var u = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    };
    if (Dn(t)) Mr(l, u);
    else {
      var n = t.alternate;
      if (t.lanes === 0 && (n === null || n.lanes === 0) && (n = l.lastRenderedReducer, n !== null))
        try {
          var i = l.lastRenderedState, f = n(i, e);
          if (u.hasEagerState = !0, u.eagerState = f, xl(f, i))
            return dn(t, l, u, 0), At === null && on(), !1;
        } catch {
        }
      if (e = Li(t, l, u, a), e !== null)
        return zl(e, t, a), Nr(e, l, a), !0;
    }
    return !1;
  }
  function Ac(t, l, e, a) {
    if (a = {
      lane: 2,
      revertLane: ef(),
      action: a,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, Dn(t)) {
      if (l) throw Error(o(479));
    } else
      l = Li(
        t,
        e,
        a,
        2
      ), l !== null && zl(l, t, 2);
  }
  function Dn(t) {
    var l = t.alternate;
    return t === lt || l !== null && l === lt;
  }
  function Mr(t, l) {
    Ca = Tn = !0;
    var e = t.pending;
    e === null ? l.next = l : (l.next = e.next, e.next = l), t.pending = l;
  }
  function Nr(t, l, e) {
    if ((e & 4194048) !== 0) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, Ke(t, e);
    }
  }
  var On = {
    readContext: al,
    use: En,
    useCallback: Ut,
    useContext: Ut,
    useEffect: Ut,
    useImperativeHandle: Ut,
    useLayoutEffect: Ut,
    useInsertionEffect: Ut,
    useMemo: Ut,
    useReducer: Ut,
    useRef: Ut,
    useState: Ut,
    useDebugValue: Ut,
    useDeferredValue: Ut,
    useTransition: Ut,
    useSyncExternalStore: Ut,
    useId: Ut,
    useHostTransitionStatus: Ut,
    useFormState: Ut,
    useActionState: Ut,
    useOptimistic: Ut,
    useMemoCache: Ut,
    useCacheRefresh: Ut
  }, Dr = {
    readContext: al,
    use: En,
    useCallback: function(t, l) {
      return ol().memoizedState = [
        t,
        l === void 0 ? null : l
      ], t;
    },
    useContext: al,
    useEffect: dr,
    useImperativeHandle: function(t, l, e) {
      e = e != null ? e.concat([t]) : null, Nn(
        4194308,
        4,
        mr.bind(null, l, t),
        e
      );
    },
    useLayoutEffect: function(t, l) {
      return Nn(4194308, 4, t, l);
    },
    useInsertionEffect: function(t, l) {
      Nn(4, 2, t, l);
    },
    useMemo: function(t, l) {
      var e = ol();
      l = l === void 0 ? null : l;
      var a = t();
      if (na) {
        il(!0);
        try {
          t();
        } finally {
          il(!1);
        }
      }
      return e.memoizedState = [a, l], a;
    },
    useReducer: function(t, l, e) {
      var a = ol();
      if (e !== void 0) {
        var u = e(l);
        if (na) {
          il(!0);
          try {
            e(l);
          } finally {
            il(!1);
          }
        }
      } else u = l;
      return a.memoizedState = a.baseState = u, t = {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: t,
        lastRenderedState: u
      }, a.queue = t, t = t.dispatch = Wh.bind(
        null,
        lt,
        t
      ), [a.memoizedState, t];
    },
    useRef: function(t) {
      var l = ol();
      return t = { current: t }, l.memoizedState = t;
    },
    useState: function(t) {
      t = Sc(t);
      var l = t.queue, e = zr.bind(null, lt, l);
      return l.dispatch = e, [t.memoizedState, e];
    },
    useDebugValue: xc,
    useDeferredValue: function(t, l) {
      var e = ol();
      return pc(e, t, l);
    },
    useTransition: function() {
      var t = Sc(!1);
      return t = pr.bind(
        null,
        lt,
        t.queue,
        !0,
        !1
      ), ol().memoizedState = t, [!1, t];
    },
    useSyncExternalStore: function(t, l, e) {
      var a = lt, u = ol();
      if (yt) {
        if (e === void 0)
          throw Error(o(407));
        e = e();
      } else {
        if (e = l(), At === null)
          throw Error(o(349));
        (ft & 124) !== 0 || Ws(a, l, e);
      }
      u.memoizedState = e;
      var n = { value: e, getSnapshot: l };
      return u.queue = n, dr(Ps.bind(null, a, n, t), [
        t
      ]), a.flags |= 2048, qa(
        9,
        Mn(),
        Fs.bind(
          null,
          a,
          n,
          e,
          l
        ),
        null
      ), e;
    },
    useId: function() {
      var t = ol(), l = At.identifierPrefix;
      if (yt) {
        var e = ie, a = ne;
        e = (a & ~(1 << 32 - ct(a) - 1)).toString(32) + e, l = "«" + l + "R" + e, e = An++, 0 < e && (l += "H" + e.toString(32)), l += "»";
      } else
        e = Vh++, l = "«" + l + "r" + e.toString(32) + "»";
      return t.memoizedState = l;
    },
    useHostTransitionStatus: Tc,
    useFormState: cr,
    useActionState: cr,
    useOptimistic: function(t) {
      var l = ol();
      l.memoizedState = l.baseState = t;
      var e = {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: null,
        lastRenderedState: null
      };
      return l.queue = e, l = Ac.bind(
        null,
        lt,
        !0,
        e
      ), e.dispatch = l, [t, l];
    },
    useMemoCache: vc,
    useCacheRefresh: function() {
      return ol().memoizedState = $h.bind(
        null,
        lt
      );
    }
  }, Or = {
    readContext: al,
    use: En,
    useCallback: Sr,
    useContext: al,
    useEffect: hr,
    useImperativeHandle: gr,
    useInsertionEffect: yr,
    useLayoutEffect: vr,
    useMemo: br,
    useReducer: zn,
    useRef: or,
    useState: function() {
      return zn(se);
    },
    useDebugValue: xc,
    useDeferredValue: function(t, l) {
      var e = Yt();
      return xr(
        e,
        gt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = zn(se)[0], l = Yt().memoizedState;
      return [
        typeof t == "boolean" ? t : Tu(t),
        l
      ];
    },
    useSyncExternalStore: $s,
    useId: Ar,
    useHostTransitionStatus: Tc,
    useFormState: fr,
    useActionState: fr,
    useOptimistic: function(t, l) {
      var e = Yt();
      return lr(e, gt, t, l);
    },
    useMemoCache: vc,
    useCacheRefresh: Er
  }, Fh = {
    readContext: al,
    use: En,
    useCallback: Sr,
    useContext: al,
    useEffect: hr,
    useImperativeHandle: gr,
    useInsertionEffect: yr,
    useLayoutEffect: vr,
    useMemo: br,
    useReducer: gc,
    useRef: or,
    useState: function() {
      return gc(se);
    },
    useDebugValue: xc,
    useDeferredValue: function(t, l) {
      var e = Yt();
      return gt === null ? pc(e, t, l) : xr(
        e,
        gt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = gc(se)[0], l = Yt().memoizedState;
      return [
        typeof t == "boolean" ? t : Tu(t),
        l
      ];
    },
    useSyncExternalStore: $s,
    useId: Ar,
    useHostTransitionStatus: Tc,
    useFormState: rr,
    useActionState: rr,
    useOptimistic: function(t, l) {
      var e = Yt();
      return gt !== null ? lr(e, gt, t, l) : (e.baseState = t, [t, e.queue.dispatch]);
    },
    useMemoCache: vc,
    useCacheRefresh: Er
  }, Ba = null, zu = 0;
  function Rn(t) {
    var l = zu;
    return zu += 1, Ba === null && (Ba = []), Qs(Ba, t, l);
  }
  function Mu(t, l) {
    l = l.props.ref, t.ref = l !== void 0 ? l : null;
  }
  function _n(t, l) {
    throw l.$$typeof === G ? Error(o(525)) : (t = Object.prototype.toString.call(l), Error(
      o(
        31,
        t === "[object Object]" ? "object with keys {" + Object.keys(l).join(", ") + "}" : t
      )
    ));
  }
  function Rr(t) {
    var l = t._init;
    return l(t._payload);
  }
  function _r(t) {
    function l(y, h) {
      if (t) {
        var v = y.deletions;
        v === null ? (y.deletions = [h], y.flags |= 16) : v.push(h);
      }
    }
    function e(y, h) {
      if (!t) return null;
      for (; h !== null; )
        l(y, h), h = h.sibling;
      return null;
    }
    function a(y) {
      for (var h = /* @__PURE__ */ new Map(); y !== null; )
        y.key !== null ? h.set(y.key, y) : h.set(y.index, y), y = y.sibling;
      return h;
    }
    function u(y, h) {
      return y = ue(y, h), y.index = 0, y.sibling = null, y;
    }
    function n(y, h, v) {
      return y.index = v, t ? (v = y.alternate, v !== null ? (v = v.index, v < h ? (y.flags |= 67108866, h) : v) : (y.flags |= 67108866, h)) : (y.flags |= 1048576, h);
    }
    function i(y) {
      return t && y.alternate === null && (y.flags |= 67108866), y;
    }
    function f(y, h, v, T) {
      return h === null || h.tag !== 6 ? (h = wi(v, y.mode, T), h.return = y, h) : (h = u(h, v), h.return = y, h);
    }
    function s(y, h, v, T) {
      var Q = v.type;
      return Q === B ? j(
        y,
        h,
        v.props.children,
        T,
        v.key
      ) : h !== null && (h.elementType === Q || typeof Q == "object" && Q !== null && Q.$$typeof === zt && Rr(Q) === h.type) ? (h = u(h, v.props), Mu(h, v), h.return = y, h) : (h = yn(
        v.type,
        v.key,
        v.props,
        null,
        y.mode,
        T
      ), Mu(h, v), h.return = y, h);
    }
    function m(y, h, v, T) {
      return h === null || h.tag !== 4 || h.stateNode.containerInfo !== v.containerInfo || h.stateNode.implementation !== v.implementation ? (h = Ki(v, y.mode, T), h.return = y, h) : (h = u(h, v.children || []), h.return = y, h);
    }
    function j(y, h, v, T, Q) {
      return h === null || h.tag !== 7 ? (h = Fe(
        v,
        y.mode,
        T,
        Q
      ), h.return = y, h) : (h = u(h, v), h.return = y, h);
    }
    function z(y, h, v) {
      if (typeof h == "string" && h !== "" || typeof h == "number" || typeof h == "bigint")
        return h = wi(
          "" + h,
          y.mode,
          v
        ), h.return = y, h;
      if (typeof h == "object" && h !== null) {
        switch (h.$$typeof) {
          case W:
            return v = yn(
              h.type,
              h.key,
              h.props,
              null,
              y.mode,
              v
            ), Mu(v, h), v.return = y, v;
          case st:
            return h = Ki(
              h,
              y.mode,
              v
            ), h.return = y, h;
          case zt:
            var T = h._init;
            return h = T(h._payload), z(y, h, v);
        }
        if (Dt(h) || Zt(h))
          return h = Fe(
            h,
            y.mode,
            v,
            null
          ), h.return = y, h;
        if (typeof h.then == "function")
          return z(y, Rn(h), v);
        if (h.$$typeof === Et)
          return z(
            y,
            Sn(y, h),
            v
          );
        _n(y, h);
      }
      return null;
    }
    function g(y, h, v, T) {
      var Q = h !== null ? h.key : null;
      if (typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint")
        return Q !== null ? null : f(y, h, "" + v, T);
      if (typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case W:
            return v.key === Q ? s(y, h, v, T) : null;
          case st:
            return v.key === Q ? m(y, h, v, T) : null;
          case zt:
            return Q = v._init, v = Q(v._payload), g(y, h, v, T);
        }
        if (Dt(v) || Zt(v))
          return Q !== null ? null : j(y, h, v, T, null);
        if (typeof v.then == "function")
          return g(
            y,
            h,
            Rn(v),
            T
          );
        if (v.$$typeof === Et)
          return g(
            y,
            h,
            Sn(y, v),
            T
          );
        _n(y, v);
      }
      return null;
    }
    function S(y, h, v, T, Q) {
      if (typeof T == "string" && T !== "" || typeof T == "number" || typeof T == "bigint")
        return y = y.get(v) || null, f(h, y, "" + T, Q);
      if (typeof T == "object" && T !== null) {
        switch (T.$$typeof) {
          case W:
            return y = y.get(
              T.key === null ? v : T.key
            ) || null, s(h, y, T, Q);
          case st:
            return y = y.get(
              T.key === null ? v : T.key
            ) || null, m(h, y, T, Q);
          case zt:
            var et = T._init;
            return T = et(T._payload), S(
              y,
              h,
              v,
              T,
              Q
            );
        }
        if (Dt(T) || Zt(T))
          return y = y.get(v) || null, j(h, y, T, Q, null);
        if (typeof T.then == "function")
          return S(
            y,
            h,
            v,
            Rn(T),
            Q
          );
        if (T.$$typeof === Et)
          return S(
            y,
            h,
            v,
            Sn(h, T),
            Q
          );
        _n(h, T);
      }
      return null;
    }
    function k(y, h, v, T) {
      for (var Q = null, et = null, Z = h, J = h = 0, kt = null; Z !== null && J < v.length; J++) {
        Z.index > J ? (kt = Z, Z = null) : kt = Z.sibling;
        var ht = g(
          y,
          Z,
          v[J],
          T
        );
        if (ht === null) {
          Z === null && (Z = kt);
          break;
        }
        t && Z && ht.alternate === null && l(y, Z), h = n(ht, h, J), et === null ? Q = ht : et.sibling = ht, et = ht, Z = kt;
      }
      if (J === v.length)
        return e(y, Z), yt && Ie(y, J), Q;
      if (Z === null) {
        for (; J < v.length; J++)
          Z = z(y, v[J], T), Z !== null && (h = n(
            Z,
            h,
            J
          ), et === null ? Q = Z : et.sibling = Z, et = Z);
        return yt && Ie(y, J), Q;
      }
      for (Z = a(Z); J < v.length; J++)
        kt = S(
          Z,
          y,
          J,
          v[J],
          T
        ), kt !== null && (t && kt.alternate !== null && Z.delete(
          kt.key === null ? J : kt.key
        ), h = n(
          kt,
          h,
          J
        ), et === null ? Q = kt : et.sibling = kt, et = kt);
      return t && Z.forEach(function(Qe) {
        return l(y, Qe);
      }), yt && Ie(y, J), Q;
    }
    function K(y, h, v, T) {
      if (v == null) throw Error(o(151));
      for (var Q = null, et = null, Z = h, J = h = 0, kt = null, ht = v.next(); Z !== null && !ht.done; J++, ht = v.next()) {
        Z.index > J ? (kt = Z, Z = null) : kt = Z.sibling;
        var Qe = g(y, Z, ht.value, T);
        if (Qe === null) {
          Z === null && (Z = kt);
          break;
        }
        t && Z && Qe.alternate === null && l(y, Z), h = n(Qe, h, J), et === null ? Q = Qe : et.sibling = Qe, et = Qe, Z = kt;
      }
      if (ht.done)
        return e(y, Z), yt && Ie(y, J), Q;
      if (Z === null) {
        for (; !ht.done; J++, ht = v.next())
          ht = z(y, ht.value, T), ht !== null && (h = n(ht, h, J), et === null ? Q = ht : et.sibling = ht, et = ht);
        return yt && Ie(y, J), Q;
      }
      for (Z = a(Z); !ht.done; J++, ht = v.next())
        ht = S(Z, y, J, ht.value, T), ht !== null && (t && ht.alternate !== null && Z.delete(ht.key === null ? J : ht.key), h = n(ht, h, J), et === null ? Q = ht : et.sibling = ht, et = ht);
      return t && Z.forEach(function(P0) {
        return l(y, P0);
      }), yt && Ie(y, J), Q;
    }
    function bt(y, h, v, T) {
      if (typeof v == "object" && v !== null && v.type === B && v.key === null && (v = v.props.children), typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case W:
            t: {
              for (var Q = v.key; h !== null; ) {
                if (h.key === Q) {
                  if (Q = v.type, Q === B) {
                    if (h.tag === 7) {
                      e(
                        y,
                        h.sibling
                      ), T = u(
                        h,
                        v.props.children
                      ), T.return = y, y = T;
                      break t;
                    }
                  } else if (h.elementType === Q || typeof Q == "object" && Q !== null && Q.$$typeof === zt && Rr(Q) === h.type) {
                    e(
                      y,
                      h.sibling
                    ), T = u(h, v.props), Mu(T, v), T.return = y, y = T;
                    break t;
                  }
                  e(y, h);
                  break;
                } else l(y, h);
                h = h.sibling;
              }
              v.type === B ? (T = Fe(
                v.props.children,
                y.mode,
                T,
                v.key
              ), T.return = y, y = T) : (T = yn(
                v.type,
                v.key,
                v.props,
                null,
                y.mode,
                T
              ), Mu(T, v), T.return = y, y = T);
            }
            return i(y);
          case st:
            t: {
              for (Q = v.key; h !== null; ) {
                if (h.key === Q)
                  if (h.tag === 4 && h.stateNode.containerInfo === v.containerInfo && h.stateNode.implementation === v.implementation) {
                    e(
                      y,
                      h.sibling
                    ), T = u(h, v.children || []), T.return = y, y = T;
                    break t;
                  } else {
                    e(y, h);
                    break;
                  }
                else l(y, h);
                h = h.sibling;
              }
              T = Ki(v, y.mode, T), T.return = y, y = T;
            }
            return i(y);
          case zt:
            return Q = v._init, v = Q(v._payload), bt(
              y,
              h,
              v,
              T
            );
        }
        if (Dt(v))
          return k(
            y,
            h,
            v,
            T
          );
        if (Zt(v)) {
          if (Q = Zt(v), typeof Q != "function") throw Error(o(150));
          return v = Q.call(v), K(
            y,
            h,
            v,
            T
          );
        }
        if (typeof v.then == "function")
          return bt(
            y,
            h,
            Rn(v),
            T
          );
        if (v.$$typeof === Et)
          return bt(
            y,
            h,
            Sn(y, v),
            T
          );
        _n(y, v);
      }
      return typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint" ? (v = "" + v, h !== null && h.tag === 6 ? (e(y, h.sibling), T = u(h, v), T.return = y, y = T) : (e(y, h), T = wi(v, y.mode, T), T.return = y, y = T), i(y)) : e(y, h);
    }
    return function(y, h, v, T) {
      try {
        zu = 0;
        var Q = bt(
          y,
          h,
          v,
          T
        );
        return Ba = null, Q;
      } catch (Z) {
        if (Z === gu || Z === xn) throw Z;
        var et = pl(29, Z, null, y.mode);
        return et.lanes = T, et.return = y, et;
      }
    };
  }
  var Ya = _r(!0), Ur = _r(!1), Cl = M(null), Jl = null;
  function ze(t) {
    var l = t.alternate;
    q(Xt, Xt.current & 1), q(Cl, t), Jl === null && (l === null || Ua.current !== null || l.memoizedState !== null) && (Jl = t);
  }
  function Cr(t) {
    if (t.tag === 22) {
      if (q(Xt, Xt.current), q(Cl, t), Jl === null) {
        var l = t.alternate;
        l !== null && l.memoizedState !== null && (Jl = t);
      }
    } else Me();
  }
  function Me() {
    q(Xt, Xt.current), q(Cl, Cl.current);
  }
  function re(t) {
    Y(Cl), Jl === t && (Jl = null), Y(Xt);
  }
  var Xt = M(0);
  function Un(t) {
    for (var l = t; l !== null; ) {
      if (l.tag === 13) {
        var e = l.memoizedState;
        if (e !== null && (e = e.dehydrated, e === null || e.data === "$?" || vf(e)))
          return l;
      } else if (l.tag === 19 && l.memoizedProps.revealOrder !== void 0) {
        if ((l.flags & 128) !== 0) return l;
      } else if (l.child !== null) {
        l.child.return = l, l = l.child;
        continue;
      }
      if (l === t) break;
      for (; l.sibling === null; ) {
        if (l.return === null || l.return === t) return null;
        l = l.return;
      }
      l.sibling.return = l.return, l = l.sibling;
    }
    return null;
  }
  function Ec(t, l, e, a) {
    l = t.memoizedState, e = e(a, l), e = e == null ? l : C({}, l, e), t.memoizedState = e, t.lanes === 0 && (t.updateQueue.baseState = e);
  }
  var zc = {
    enqueueSetState: function(t, l, e) {
      t = t._reactInternals;
      var a = El(), u = Te(a);
      u.payload = l, e != null && (u.callback = e), l = Ae(t, u, a), l !== null && (zl(l, t, a), bu(l, t, a));
    },
    enqueueReplaceState: function(t, l, e) {
      t = t._reactInternals;
      var a = El(), u = Te(a);
      u.tag = 1, u.payload = l, e != null && (u.callback = e), l = Ae(t, u, a), l !== null && (zl(l, t, a), bu(l, t, a));
    },
    enqueueForceUpdate: function(t, l) {
      t = t._reactInternals;
      var e = El(), a = Te(e);
      a.tag = 2, l != null && (a.callback = l), l = Ae(t, a, e), l !== null && (zl(l, t, e), bu(l, t, e));
    }
  };
  function Hr(t, l, e, a, u, n, i) {
    return t = t.stateNode, typeof t.shouldComponentUpdate == "function" ? t.shouldComponentUpdate(a, n, i) : l.prototype && l.prototype.isPureReactComponent ? !su(e, a) || !su(u, n) : !0;
  }
  function qr(t, l, e, a) {
    t = l.state, typeof l.componentWillReceiveProps == "function" && l.componentWillReceiveProps(e, a), typeof l.UNSAFE_componentWillReceiveProps == "function" && l.UNSAFE_componentWillReceiveProps(e, a), l.state !== t && zc.enqueueReplaceState(l, l.state, null);
  }
  function ia(t, l) {
    var e = l;
    if ("ref" in l) {
      e = {};
      for (var a in l)
        a !== "ref" && (e[a] = l[a]);
    }
    if (t = t.defaultProps) {
      e === l && (e = C({}, e));
      for (var u in t)
        e[u] === void 0 && (e[u] = t[u]);
    }
    return e;
  }
  var Cn = typeof reportError == "function" ? reportError : function(t) {
    if (typeof window == "object" && typeof window.ErrorEvent == "function") {
      var l = new window.ErrorEvent("error", {
        bubbles: !0,
        cancelable: !0,
        message: typeof t == "object" && t !== null && typeof t.message == "string" ? String(t.message) : String(t),
        error: t
      });
      if (!window.dispatchEvent(l)) return;
    } else if (typeof process == "object" && typeof process.emit == "function") {
      process.emit("uncaughtException", t);
      return;
    }
    console.error(t);
  };
  function Br(t) {
    Cn(t);
  }
  function Yr(t) {
    console.error(t);
  }
  function Gr(t) {
    Cn(t);
  }
  function Hn(t, l) {
    try {
      var e = t.onUncaughtError;
      e(l.value, { componentStack: l.stack });
    } catch (a) {
      setTimeout(function() {
        throw a;
      });
    }
  }
  function Xr(t, l, e) {
    try {
      var a = t.onCaughtError;
      a(e.value, {
        componentStack: e.stack,
        errorBoundary: l.tag === 1 ? l.stateNode : null
      });
    } catch (u) {
      setTimeout(function() {
        throw u;
      });
    }
  }
  function Mc(t, l, e) {
    return e = Te(e), e.tag = 3, e.payload = { element: null }, e.callback = function() {
      Hn(t, l);
    }, e;
  }
  function Qr(t) {
    return t = Te(t), t.tag = 3, t;
  }
  function Zr(t, l, e, a) {
    var u = e.type.getDerivedStateFromError;
    if (typeof u == "function") {
      var n = a.value;
      t.payload = function() {
        return u(n);
      }, t.callback = function() {
        Xr(l, e, a);
      };
    }
    var i = e.stateNode;
    i !== null && typeof i.componentDidCatch == "function" && (t.callback = function() {
      Xr(l, e, a), typeof u != "function" && (Ue === null ? Ue = /* @__PURE__ */ new Set([this]) : Ue.add(this));
      var f = a.stack;
      this.componentDidCatch(a.value, {
        componentStack: f !== null ? f : ""
      });
    });
  }
  function Ph(t, l, e, a, u) {
    if (e.flags |= 32768, a !== null && typeof a == "object" && typeof a.then == "function") {
      if (l = e.alternate, l !== null && yu(
        l,
        e,
        u,
        !0
      ), e = Cl.current, e !== null) {
        switch (e.tag) {
          case 13:
            return Jl === null ? Fc() : e.alternate === null && Rt === 0 && (Rt = 3), e.flags &= -257, e.flags |= 65536, e.lanes = u, a === ec ? e.flags |= 16384 : (l = e.updateQueue, l === null ? e.updateQueue = /* @__PURE__ */ new Set([a]) : l.add(a), Ic(t, a, u)), !1;
          case 22:
            return e.flags |= 65536, a === ec ? e.flags |= 16384 : (l = e.updateQueue, l === null ? (l = {
              transitions: null,
              markerInstances: null,
              retryQueue: /* @__PURE__ */ new Set([a])
            }, e.updateQueue = l) : (e = l.retryQueue, e === null ? l.retryQueue = /* @__PURE__ */ new Set([a]) : e.add(a)), Ic(t, a, u)), !1;
        }
        throw Error(o(435, e.tag));
      }
      return Ic(t, a, u), Fc(), !1;
    }
    if (yt)
      return l = Cl.current, l !== null ? ((l.flags & 65536) === 0 && (l.flags |= 256), l.flags |= 65536, l.lanes = u, a !== $i && (t = Error(o(422), { cause: a }), hu(Ol(t, e)))) : (a !== $i && (l = Error(o(423), {
        cause: a
      }), hu(
        Ol(l, e)
      )), t = t.current.alternate, t.flags |= 65536, u &= -u, t.lanes |= u, a = Ol(a, e), u = Mc(
        t.stateNode,
        a,
        u
      ), nc(t, u), Rt !== 4 && (Rt = 2)), !1;
    var n = Error(o(520), { cause: a });
    if (n = Ol(n, e), Cu === null ? Cu = [n] : Cu.push(n), Rt !== 4 && (Rt = 2), l === null) return !0;
    a = Ol(a, e), e = l;
    do {
      switch (e.tag) {
        case 3:
          return e.flags |= 65536, t = u & -u, e.lanes |= t, t = Mc(e.stateNode, a, t), nc(e, t), !1;
        case 1:
          if (l = e.type, n = e.stateNode, (e.flags & 128) === 0 && (typeof l.getDerivedStateFromError == "function" || n !== null && typeof n.componentDidCatch == "function" && (Ue === null || !Ue.has(n))))
            return e.flags |= 65536, u &= -u, e.lanes |= u, u = Qr(u), Zr(
              u,
              t,
              e,
              a
            ), nc(e, u), !1;
      }
      e = e.return;
    } while (e !== null);
    return !1;
  }
  var Lr = Error(o(461)), Kt = !1;
  function Pt(t, l, e, a) {
    l.child = t === null ? Ur(l, null, e, a) : Ya(
      l,
      t.child,
      e,
      a
    );
  }
  function Vr(t, l, e, a, u) {
    e = e.render;
    var n = l.ref;
    if ("ref" in a) {
      var i = {};
      for (var f in a)
        f !== "ref" && (i[f] = a[f]);
    } else i = a;
    return aa(l), a = rc(
      t,
      l,
      e,
      i,
      n,
      u
    ), f = oc(), t !== null && !Kt ? (dc(t, l, u), oe(t, l, u)) : (yt && f && Ji(l), l.flags |= 1, Pt(t, l, a, u), l.child);
  }
  function wr(t, l, e, a, u) {
    if (t === null) {
      var n = e.type;
      return typeof n == "function" && !Vi(n) && n.defaultProps === void 0 && e.compare === null ? (l.tag = 15, l.type = n, Kr(
        t,
        l,
        n,
        a,
        u
      )) : (t = yn(
        e.type,
        null,
        a,
        l,
        l.mode,
        u
      ), t.ref = l.ref, t.return = l, l.child = t);
    }
    if (n = t.child, !Hc(t, u)) {
      var i = n.memoizedProps;
      if (e = e.compare, e = e !== null ? e : su, e(i, a) && t.ref === l.ref)
        return oe(t, l, u);
    }
    return l.flags |= 1, t = ue(n, a), t.ref = l.ref, t.return = l, l.child = t;
  }
  function Kr(t, l, e, a, u) {
    if (t !== null) {
      var n = t.memoizedProps;
      if (su(n, a) && t.ref === l.ref)
        if (Kt = !1, l.pendingProps = a = n, Hc(t, u))
          (t.flags & 131072) !== 0 && (Kt = !0);
        else
          return l.lanes = t.lanes, oe(t, l, u);
    }
    return Nc(
      t,
      l,
      e,
      a,
      u
    );
  }
  function Jr(t, l, e) {
    var a = l.pendingProps, u = a.children, n = t !== null ? t.memoizedState : null;
    if (a.mode === "hidden") {
      if ((l.flags & 128) !== 0) {
        if (a = n !== null ? n.baseLanes | e : e, t !== null) {
          for (u = l.child = t.child, n = 0; u !== null; )
            n = n | u.lanes | u.childLanes, u = u.sibling;
          l.childLanes = n & ~a;
        } else l.childLanes = 0, l.child = null;
        return kr(
          t,
          l,
          a,
          e
        );
      }
      if ((e & 536870912) !== 0)
        l.memoizedState = { baseLanes: 0, cachePool: null }, t !== null && bn(
          l,
          n !== null ? n.cachePool : null
        ), n !== null ? Ks(l, n) : cc(), Cr(l);
      else
        return l.lanes = l.childLanes = 536870912, kr(
          t,
          l,
          n !== null ? n.baseLanes | e : e,
          e
        );
    } else
      n !== null ? (bn(l, n.cachePool), Ks(l, n), Me(), l.memoizedState = null) : (t !== null && bn(l, null), cc(), Me());
    return Pt(t, l, u, e), l.child;
  }
  function kr(t, l, e, a) {
    var u = lc();
    return u = u === null ? null : { parent: Gt._currentValue, pool: u }, l.memoizedState = {
      baseLanes: e,
      cachePool: u
    }, t !== null && bn(l, null), cc(), Cr(l), t !== null && yu(t, l, a, !0), null;
  }
  function qn(t, l) {
    var e = l.ref;
    if (e === null)
      t !== null && t.ref !== null && (l.flags |= 4194816);
    else {
      if (typeof e != "function" && typeof e != "object")
        throw Error(o(284));
      (t === null || t.ref !== e) && (l.flags |= 4194816);
    }
  }
  function Nc(t, l, e, a, u) {
    return aa(l), e = rc(
      t,
      l,
      e,
      a,
      void 0,
      u
    ), a = oc(), t !== null && !Kt ? (dc(t, l, u), oe(t, l, u)) : (yt && a && Ji(l), l.flags |= 1, Pt(t, l, e, u), l.child);
  }
  function $r(t, l, e, a, u, n) {
    return aa(l), l.updateQueue = null, e = ks(
      l,
      a,
      e,
      u
    ), Js(t), a = oc(), t !== null && !Kt ? (dc(t, l, n), oe(t, l, n)) : (yt && a && Ji(l), l.flags |= 1, Pt(t, l, e, n), l.child);
  }
  function Wr(t, l, e, a, u) {
    if (aa(l), l.stateNode === null) {
      var n = Na, i = e.contextType;
      typeof i == "object" && i !== null && (n = al(i)), n = new e(a, n), l.memoizedState = n.state !== null && n.state !== void 0 ? n.state : null, n.updater = zc, l.stateNode = n, n._reactInternals = l, n = l.stateNode, n.props = a, n.state = l.memoizedState, n.refs = {}, ac(l), i = e.contextType, n.context = typeof i == "object" && i !== null ? al(i) : Na, n.state = l.memoizedState, i = e.getDerivedStateFromProps, typeof i == "function" && (Ec(
        l,
        e,
        i,
        a
      ), n.state = l.memoizedState), typeof e.getDerivedStateFromProps == "function" || typeof n.getSnapshotBeforeUpdate == "function" || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (i = n.state, typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount(), i !== n.state && zc.enqueueReplaceState(n, n.state, null), pu(l, a, n, u), xu(), n.state = l.memoizedState), typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !0;
    } else if (t === null) {
      n = l.stateNode;
      var f = l.memoizedProps, s = ia(e, f);
      n.props = s;
      var m = n.context, j = e.contextType;
      i = Na, typeof j == "object" && j !== null && (i = al(j));
      var z = e.getDerivedStateFromProps;
      j = typeof z == "function" || typeof n.getSnapshotBeforeUpdate == "function", f = l.pendingProps !== f, j || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (f || m !== i) && qr(
        l,
        n,
        a,
        i
      ), je = !1;
      var g = l.memoizedState;
      n.state = g, pu(l, a, n, u), xu(), m = l.memoizedState, f || g !== m || je ? (typeof z == "function" && (Ec(
        l,
        e,
        z,
        a
      ), m = l.memoizedState), (s = je || Hr(
        l,
        e,
        s,
        a,
        g,
        m,
        i
      )) ? (j || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount()), typeof n.componentDidMount == "function" && (l.flags |= 4194308)) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), l.memoizedProps = a, l.memoizedState = m), n.props = a, n.state = m, n.context = i, a = s) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !1);
    } else {
      n = l.stateNode, uc(t, l), i = l.memoizedProps, j = ia(e, i), n.props = j, z = l.pendingProps, g = n.context, m = e.contextType, s = Na, typeof m == "object" && m !== null && (s = al(m)), f = e.getDerivedStateFromProps, (m = typeof f == "function" || typeof n.getSnapshotBeforeUpdate == "function") || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (i !== z || g !== s) && qr(
        l,
        n,
        a,
        s
      ), je = !1, g = l.memoizedState, n.state = g, pu(l, a, n, u), xu();
      var S = l.memoizedState;
      i !== z || g !== S || je || t !== null && t.dependencies !== null && gn(t.dependencies) ? (typeof f == "function" && (Ec(
        l,
        e,
        f,
        a
      ), S = l.memoizedState), (j = je || Hr(
        l,
        e,
        j,
        a,
        g,
        S,
        s
      ) || t !== null && t.dependencies !== null && gn(t.dependencies)) ? (m || typeof n.UNSAFE_componentWillUpdate != "function" && typeof n.componentWillUpdate != "function" || (typeof n.componentWillUpdate == "function" && n.componentWillUpdate(a, S, s), typeof n.UNSAFE_componentWillUpdate == "function" && n.UNSAFE_componentWillUpdate(
        a,
        S,
        s
      )), typeof n.componentDidUpdate == "function" && (l.flags |= 4), typeof n.getSnapshotBeforeUpdate == "function" && (l.flags |= 1024)) : (typeof n.componentDidUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), l.memoizedProps = a, l.memoizedState = S), n.props = a, n.state = S, n.context = s, a = j) : (typeof n.componentDidUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), a = !1);
    }
    return n = a, qn(t, l), a = (l.flags & 128) !== 0, n || a ? (n = l.stateNode, e = a && typeof e.getDerivedStateFromError != "function" ? null : n.render(), l.flags |= 1, t !== null && a ? (l.child = Ya(
      l,
      t.child,
      null,
      u
    ), l.child = Ya(
      l,
      null,
      e,
      u
    )) : Pt(t, l, e, u), l.memoizedState = n.state, t = l.child) : t = oe(
      t,
      l,
      u
    ), t;
  }
  function Fr(t, l, e, a) {
    return du(), l.flags |= 256, Pt(t, l, e, a), l.child;
  }
  var Dc = {
    dehydrated: null,
    treeContext: null,
    retryLane: 0,
    hydrationErrors: null
  };
  function Oc(t) {
    return { baseLanes: t, cachePool: Ys() };
  }
  function Rc(t, l, e) {
    return t = t !== null ? t.childLanes & ~e : 0, l && (t |= Hl), t;
  }
  function Pr(t, l, e) {
    var a = l.pendingProps, u = !1, n = (l.flags & 128) !== 0, i;
    if ((i = n) || (i = t !== null && t.memoizedState === null ? !1 : (Xt.current & 2) !== 0), i && (u = !0, l.flags &= -129), i = (l.flags & 32) !== 0, l.flags &= -33, t === null) {
      if (yt) {
        if (u ? ze(l) : Me(), yt) {
          var f = Ot, s;
          if (s = f) {
            t: {
              for (s = f, f = Kl; s.nodeType !== 8; ) {
                if (!f) {
                  f = null;
                  break t;
                }
                if (s = Xl(
                  s.nextSibling
                ), s === null) {
                  f = null;
                  break t;
                }
              }
              f = s;
            }
            f !== null ? (l.memoizedState = {
              dehydrated: f,
              treeContext: Pe !== null ? { id: ne, overflow: ie } : null,
              retryLane: 536870912,
              hydrationErrors: null
            }, s = pl(
              18,
              null,
              null,
              0
            ), s.stateNode = f, s.return = l, l.child = s, fl = l, Ot = null, s = !0) : s = !1;
          }
          s || la(l);
        }
        if (f = l.memoizedState, f !== null && (f = f.dehydrated, f !== null))
          return vf(f) ? l.lanes = 32 : l.lanes = 536870912, null;
        re(l);
      }
      return f = a.children, a = a.fallback, u ? (Me(), u = l.mode, f = Bn(
        { mode: "hidden", children: f },
        u
      ), a = Fe(
        a,
        u,
        e,
        null
      ), f.return = l, a.return = l, f.sibling = a, l.child = f, u = l.child, u.memoizedState = Oc(e), u.childLanes = Rc(
        t,
        i,
        e
      ), l.memoizedState = Dc, a) : (ze(l), _c(l, f));
    }
    if (s = t.memoizedState, s !== null && (f = s.dehydrated, f !== null)) {
      if (n)
        l.flags & 256 ? (ze(l), l.flags &= -257, l = Uc(
          t,
          l,
          e
        )) : l.memoizedState !== null ? (Me(), l.child = t.child, l.flags |= 128, l = null) : (Me(), u = a.fallback, f = l.mode, a = Bn(
          { mode: "visible", children: a.children },
          f
        ), u = Fe(
          u,
          f,
          e,
          null
        ), u.flags |= 2, a.return = l, u.return = l, a.sibling = u, l.child = a, Ya(
          l,
          t.child,
          null,
          e
        ), a = l.child, a.memoizedState = Oc(e), a.childLanes = Rc(
          t,
          i,
          e
        ), l.memoizedState = Dc, l = u);
      else if (ze(l), vf(f)) {
        if (i = f.nextSibling && f.nextSibling.dataset, i) var m = i.dgst;
        i = m, a = Error(o(419)), a.stack = "", a.digest = i, hu({ value: a, source: null, stack: null }), l = Uc(
          t,
          l,
          e
        );
      } else if (Kt || yu(t, l, e, !1), i = (e & t.childLanes) !== 0, Kt || i) {
        if (i = At, i !== null && (a = e & -e, a = (a & 42) !== 0 ? 1 : cl(a), a = (a & (i.suspendedLanes | e)) !== 0 ? 0 : a, a !== 0 && a !== s.retryLane))
          throw s.retryLane = a, Ma(t, a), zl(i, t, a), Lr;
        f.data === "$?" || Fc(), l = Uc(
          t,
          l,
          e
        );
      } else
        f.data === "$?" ? (l.flags |= 192, l.child = t.child, l = null) : (t = s.treeContext, Ot = Xl(
          f.nextSibling
        ), fl = l, yt = !0, ta = null, Kl = !1, t !== null && (_l[Ul++] = ne, _l[Ul++] = ie, _l[Ul++] = Pe, ne = t.id, ie = t.overflow, Pe = l), l = _c(
          l,
          a.children
        ), l.flags |= 4096);
      return l;
    }
    return u ? (Me(), u = a.fallback, f = l.mode, s = t.child, m = s.sibling, a = ue(s, {
      mode: "hidden",
      children: a.children
    }), a.subtreeFlags = s.subtreeFlags & 65011712, m !== null ? u = ue(m, u) : (u = Fe(
      u,
      f,
      e,
      null
    ), u.flags |= 2), u.return = l, a.return = l, a.sibling = u, l.child = a, a = u, u = l.child, f = t.child.memoizedState, f === null ? f = Oc(e) : (s = f.cachePool, s !== null ? (m = Gt._currentValue, s = s.parent !== m ? { parent: m, pool: m } : s) : s = Ys(), f = {
      baseLanes: f.baseLanes | e,
      cachePool: s
    }), u.memoizedState = f, u.childLanes = Rc(
      t,
      i,
      e
    ), l.memoizedState = Dc, a) : (ze(l), e = t.child, t = e.sibling, e = ue(e, {
      mode: "visible",
      children: a.children
    }), e.return = l, e.sibling = null, t !== null && (i = l.deletions, i === null ? (l.deletions = [t], l.flags |= 16) : i.push(t)), l.child = e, l.memoizedState = null, e);
  }
  function _c(t, l) {
    return l = Bn(
      { mode: "visible", children: l },
      t.mode
    ), l.return = t, t.child = l;
  }
  function Bn(t, l) {
    return t = pl(22, t, null, l), t.lanes = 0, t.stateNode = {
      _visibility: 1,
      _pendingMarkers: null,
      _retryCache: null,
      _transitions: null
    }, t;
  }
  function Uc(t, l, e) {
    return Ya(l, t.child, null, e), t = _c(
      l,
      l.pendingProps.children
    ), t.flags |= 2, l.memoizedState = null, t;
  }
  function Ir(t, l, e) {
    t.lanes |= l;
    var a = t.alternate;
    a !== null && (a.lanes |= l), Fi(t.return, l, e);
  }
  function Cc(t, l, e, a, u) {
    var n = t.memoizedState;
    n === null ? t.memoizedState = {
      isBackwards: l,
      rendering: null,
      renderingStartTime: 0,
      last: a,
      tail: e,
      tailMode: u
    } : (n.isBackwards = l, n.rendering = null, n.renderingStartTime = 0, n.last = a, n.tail = e, n.tailMode = u);
  }
  function to(t, l, e) {
    var a = l.pendingProps, u = a.revealOrder, n = a.tail;
    if (Pt(t, l, a.children, e), a = Xt.current, (a & 2) !== 0)
      a = a & 1 | 2, l.flags |= 128;
    else {
      if (t !== null && (t.flags & 128) !== 0)
        t: for (t = l.child; t !== null; ) {
          if (t.tag === 13)
            t.memoizedState !== null && Ir(t, e, l);
          else if (t.tag === 19)
            Ir(t, e, l);
          else if (t.child !== null) {
            t.child.return = t, t = t.child;
            continue;
          }
          if (t === l) break t;
          for (; t.sibling === null; ) {
            if (t.return === null || t.return === l)
              break t;
            t = t.return;
          }
          t.sibling.return = t.return, t = t.sibling;
        }
      a &= 1;
    }
    switch (q(Xt, a), u) {
      case "forwards":
        for (e = l.child, u = null; e !== null; )
          t = e.alternate, t !== null && Un(t) === null && (u = e), e = e.sibling;
        e = u, e === null ? (u = l.child, l.child = null) : (u = e.sibling, e.sibling = null), Cc(
          l,
          !1,
          u,
          e,
          n
        );
        break;
      case "backwards":
        for (e = null, u = l.child, l.child = null; u !== null; ) {
          if (t = u.alternate, t !== null && Un(t) === null) {
            l.child = u;
            break;
          }
          t = u.sibling, u.sibling = e, e = u, u = t;
        }
        Cc(
          l,
          !0,
          e,
          null,
          n
        );
        break;
      case "together":
        Cc(l, !1, null, null, void 0);
        break;
      default:
        l.memoizedState = null;
    }
    return l.child;
  }
  function oe(t, l, e) {
    if (t !== null && (l.dependencies = t.dependencies), _e |= l.lanes, (e & l.childLanes) === 0)
      if (t !== null) {
        if (yu(
          t,
          l,
          e,
          !1
        ), (e & l.childLanes) === 0)
          return null;
      } else return null;
    if (t !== null && l.child !== t.child)
      throw Error(o(153));
    if (l.child !== null) {
      for (t = l.child, e = ue(t, t.pendingProps), l.child = e, e.return = l; t.sibling !== null; )
        t = t.sibling, e = e.sibling = ue(t, t.pendingProps), e.return = l;
      e.sibling = null;
    }
    return l.child;
  }
  function Hc(t, l) {
    return (t.lanes & l) !== 0 ? !0 : (t = t.dependencies, !!(t !== null && gn(t)));
  }
  function Ih(t, l, e) {
    switch (l.tag) {
      case 3:
        Tt(l, l.stateNode.containerInfo), pe(l, Gt, t.memoizedState.cache), du();
        break;
      case 27:
      case 5:
        Pa(l);
        break;
      case 4:
        Tt(l, l.stateNode.containerInfo);
        break;
      case 10:
        pe(
          l,
          l.type,
          l.memoizedProps.value
        );
        break;
      case 13:
        var a = l.memoizedState;
        if (a !== null)
          return a.dehydrated !== null ? (ze(l), l.flags |= 128, null) : (e & l.child.childLanes) !== 0 ? Pr(t, l, e) : (ze(l), t = oe(
            t,
            l,
            e
          ), t !== null ? t.sibling : null);
        ze(l);
        break;
      case 19:
        var u = (t.flags & 128) !== 0;
        if (a = (e & l.childLanes) !== 0, a || (yu(
          t,
          l,
          e,
          !1
        ), a = (e & l.childLanes) !== 0), u) {
          if (a)
            return to(
              t,
              l,
              e
            );
          l.flags |= 128;
        }
        if (u = l.memoizedState, u !== null && (u.rendering = null, u.tail = null, u.lastEffect = null), q(Xt, Xt.current), a) break;
        return null;
      case 22:
      case 23:
        return l.lanes = 0, Jr(t, l, e);
      case 24:
        pe(l, Gt, t.memoizedState.cache);
    }
    return oe(t, l, e);
  }
  function lo(t, l, e) {
    if (t !== null)
      if (t.memoizedProps !== l.pendingProps)
        Kt = !0;
      else {
        if (!Hc(t, e) && (l.flags & 128) === 0)
          return Kt = !1, Ih(
            t,
            l,
            e
          );
        Kt = (t.flags & 131072) !== 0;
      }
    else
      Kt = !1, yt && (l.flags & 1048576) !== 0 && Rs(l, mn, l.index);
    switch (l.lanes = 0, l.tag) {
      case 16:
        t: {
          t = l.pendingProps;
          var a = l.elementType, u = a._init;
          if (a = u(a._payload), l.type = a, typeof a == "function")
            Vi(a) ? (t = ia(a, t), l.tag = 1, l = Wr(
              null,
              l,
              a,
              t,
              e
            )) : (l.tag = 0, l = Nc(
              null,
              l,
              a,
              t,
              e
            ));
          else {
            if (a != null) {
              if (u = a.$$typeof, u === nt) {
                l.tag = 11, l = Vr(
                  null,
                  l,
                  a,
                  t,
                  e
                );
                break t;
              } else if (u === V) {
                l.tag = 14, l = wr(
                  null,
                  l,
                  a,
                  t,
                  e
                );
                break t;
              }
            }
            throw l = vl(a) || a, Error(o(306, l, ""));
          }
        }
        return l;
      case 0:
        return Nc(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 1:
        return a = l.type, u = ia(
          a,
          l.pendingProps
        ), Wr(
          t,
          l,
          a,
          u,
          e
        );
      case 3:
        t: {
          if (Tt(
            l,
            l.stateNode.containerInfo
          ), t === null) throw Error(o(387));
          a = l.pendingProps;
          var n = l.memoizedState;
          u = n.element, uc(t, l), pu(l, a, null, e);
          var i = l.memoizedState;
          if (a = i.cache, pe(l, Gt, a), a !== n.cache && Pi(
            l,
            [Gt],
            e,
            !0
          ), xu(), a = i.element, n.isDehydrated)
            if (n = {
              element: a,
              isDehydrated: !1,
              cache: i.cache
            }, l.updateQueue.baseState = n, l.memoizedState = n, l.flags & 256) {
              l = Fr(
                t,
                l,
                a,
                e
              );
              break t;
            } else if (a !== u) {
              u = Ol(
                Error(o(424)),
                l
              ), hu(u), l = Fr(
                t,
                l,
                a,
                e
              );
              break t;
            } else
              for (t = l.stateNode.containerInfo, t.nodeType === 9 ? t = t.body : t = t.nodeName === "HTML" ? t.ownerDocument.body : t, Ot = Xl(t.firstChild), fl = l, yt = !0, ta = null, Kl = !0, e = Ur(
                l,
                null,
                a,
                e
              ), l.child = e; e; )
                e.flags = e.flags & -3 | 4096, e = e.sibling;
          else {
            if (du(), a === u) {
              l = oe(
                t,
                l,
                e
              );
              break t;
            }
            Pt(
              t,
              l,
              a,
              e
            );
          }
          l = l.child;
        }
        return l;
      case 26:
        return qn(t, l), t === null ? (e = id(
          l.type,
          null,
          l.pendingProps,
          null
        )) ? l.memoizedState = e : yt || (e = l.type, t = l.pendingProps, a = Fn(
          $.current
        ).createElement(e), a[el] = l, a[sl] = t, tl(a, e, t), wt(a), l.stateNode = a) : l.memoizedState = id(
          l.type,
          t.memoizedProps,
          l.pendingProps,
          t.memoizedState
        ), null;
      case 27:
        return Pa(l), t === null && yt && (a = l.stateNode = ad(
          l.type,
          l.pendingProps,
          $.current
        ), fl = l, Kl = !0, u = Ot, qe(l.type) ? (mf = u, Ot = Xl(
          a.firstChild
        )) : Ot = u), Pt(
          t,
          l,
          l.pendingProps.children,
          e
        ), qn(t, l), t === null && (l.flags |= 4194304), l.child;
      case 5:
        return t === null && yt && ((u = a = Ot) && (a = M0(
          a,
          l.type,
          l.pendingProps,
          Kl
        ), a !== null ? (l.stateNode = a, fl = l, Ot = Xl(
          a.firstChild
        ), Kl = !1, u = !0) : u = !1), u || la(l)), Pa(l), u = l.type, n = l.pendingProps, i = t !== null ? t.memoizedProps : null, a = n.children, df(u, n) ? a = null : i !== null && df(u, i) && (l.flags |= 32), l.memoizedState !== null && (u = rc(
          t,
          l,
          wh,
          null,
          null,
          e
        ), Lu._currentValue = u), qn(t, l), Pt(t, l, a, e), l.child;
      case 6:
        return t === null && yt && ((t = e = Ot) && (e = N0(
          e,
          l.pendingProps,
          Kl
        ), e !== null ? (l.stateNode = e, fl = l, Ot = null, t = !0) : t = !1), t || la(l)), null;
      case 13:
        return Pr(t, l, e);
      case 4:
        return Tt(
          l,
          l.stateNode.containerInfo
        ), a = l.pendingProps, t === null ? l.child = Ya(
          l,
          null,
          a,
          e
        ) : Pt(
          t,
          l,
          a,
          e
        ), l.child;
      case 11:
        return Vr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 7:
        return Pt(
          t,
          l,
          l.pendingProps,
          e
        ), l.child;
      case 8:
        return Pt(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 12:
        return Pt(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 10:
        return a = l.pendingProps, pe(l, l.type, a.value), Pt(
          t,
          l,
          a.children,
          e
        ), l.child;
      case 9:
        return u = l.type._context, a = l.pendingProps.children, aa(l), u = al(u), a = a(u), l.flags |= 1, Pt(t, l, a, e), l.child;
      case 14:
        return wr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 15:
        return Kr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 19:
        return to(t, l, e);
      case 31:
        return a = l.pendingProps, e = l.mode, a = {
          mode: a.mode,
          children: a.children
        }, t === null ? (e = Bn(
          a,
          e
        ), e.ref = l.ref, l.child = e, e.return = l, l = e) : (e = ue(t.child, a), e.ref = l.ref, l.child = e, e.return = l, l = e), l;
      case 22:
        return Jr(t, l, e);
      case 24:
        return aa(l), a = al(Gt), t === null ? (u = lc(), u === null && (u = At, n = Ii(), u.pooledCache = n, n.refCount++, n !== null && (u.pooledCacheLanes |= e), u = n), l.memoizedState = {
          parent: a,
          cache: u
        }, ac(l), pe(l, Gt, u)) : ((t.lanes & e) !== 0 && (uc(t, l), pu(l, null, null, e), xu()), u = t.memoizedState, n = l.memoizedState, u.parent !== a ? (u = { parent: a, cache: a }, l.memoizedState = u, l.lanes === 0 && (l.memoizedState = l.updateQueue.baseState = u), pe(l, Gt, a)) : (a = n.cache, pe(l, Gt, a), a !== u.cache && Pi(
          l,
          [Gt],
          e,
          !0
        ))), Pt(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 29:
        throw l.pendingProps;
    }
    throw Error(o(156, l.tag));
  }
  function de(t) {
    t.flags |= 4;
  }
  function eo(t, l) {
    if (l.type !== "stylesheet" || (l.state.loading & 4) !== 0)
      t.flags &= -16777217;
    else if (t.flags |= 16777216, !od(l)) {
      if (l = Cl.current, l !== null && ((ft & 4194048) === ft ? Jl !== null : (ft & 62914560) !== ft && (ft & 536870912) === 0 || l !== Jl))
        throw Su = ec, Gs;
      t.flags |= 8192;
    }
  }
  function Yn(t, l) {
    l !== null && (t.flags |= 4), t.flags & 16384 && (l = t.tag !== 22 ? X() : 536870912, t.lanes |= l, Za |= l);
  }
  function Nu(t, l) {
    if (!yt)
      switch (t.tailMode) {
        case "hidden":
          l = t.tail;
          for (var e = null; l !== null; )
            l.alternate !== null && (e = l), l = l.sibling;
          e === null ? t.tail = null : e.sibling = null;
          break;
        case "collapsed":
          e = t.tail;
          for (var a = null; e !== null; )
            e.alternate !== null && (a = e), e = e.sibling;
          a === null ? l || t.tail === null ? t.tail = null : t.tail.sibling = null : a.sibling = null;
      }
  }
  function Nt(t) {
    var l = t.alternate !== null && t.alternate.child === t.child, e = 0, a = 0;
    if (l)
      for (var u = t.child; u !== null; )
        e |= u.lanes | u.childLanes, a |= u.subtreeFlags & 65011712, a |= u.flags & 65011712, u.return = t, u = u.sibling;
    else
      for (u = t.child; u !== null; )
        e |= u.lanes | u.childLanes, a |= u.subtreeFlags, a |= u.flags, u.return = t, u = u.sibling;
    return t.subtreeFlags |= a, t.childLanes = e, l;
  }
  function t0(t, l, e) {
    var a = l.pendingProps;
    switch (ki(l), l.tag) {
      case 31:
      case 16:
      case 15:
      case 0:
      case 11:
      case 7:
      case 8:
      case 12:
      case 9:
      case 14:
        return Nt(l), null;
      case 1:
        return Nt(l), null;
      case 3:
        return e = l.stateNode, a = null, t !== null && (a = t.memoizedState.cache), l.memoizedState.cache !== a && (l.flags |= 2048), fe(Gt), Zl(), e.pendingContext && (e.context = e.pendingContext, e.pendingContext = null), (t === null || t.child === null) && (ou(l) ? de(l) : t === null || t.memoizedState.isDehydrated && (l.flags & 256) === 0 || (l.flags |= 1024, Cs())), Nt(l), null;
      case 26:
        return e = l.memoizedState, t === null ? (de(l), e !== null ? (Nt(l), eo(l, e)) : (Nt(l), l.flags &= -16777217)) : e ? e !== t.memoizedState ? (de(l), Nt(l), eo(l, e)) : (Nt(l), l.flags &= -16777217) : (t.memoizedProps !== a && de(l), Nt(l), l.flags &= -16777217), null;
      case 27:
        Se(l), e = $.current;
        var u = l.type;
        if (t !== null && l.stateNode != null)
          t.memoizedProps !== a && de(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(o(166));
            return Nt(l), null;
          }
          t = L.current, ou(l) ? _s(l) : (t = ad(u, a, e), l.stateNode = t, de(l));
        }
        return Nt(l), null;
      case 5:
        if (Se(l), e = l.type, t !== null && l.stateNode != null)
          t.memoizedProps !== a && de(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(o(166));
            return Nt(l), null;
          }
          if (t = L.current, ou(l))
            _s(l);
          else {
            switch (u = Fn(
              $.current
            ), t) {
              case 1:
                t = u.createElementNS(
                  "http://www.w3.org/2000/svg",
                  e
                );
                break;
              case 2:
                t = u.createElementNS(
                  "http://www.w3.org/1998/Math/MathML",
                  e
                );
                break;
              default:
                switch (e) {
                  case "svg":
                    t = u.createElementNS(
                      "http://www.w3.org/2000/svg",
                      e
                    );
                    break;
                  case "math":
                    t = u.createElementNS(
                      "http://www.w3.org/1998/Math/MathML",
                      e
                    );
                    break;
                  case "script":
                    t = u.createElement("div"), t.innerHTML = "<script><\/script>", t = t.removeChild(t.firstChild);
                    break;
                  case "select":
                    t = typeof a.is == "string" ? u.createElement("select", { is: a.is }) : u.createElement("select"), a.multiple ? t.multiple = !0 : a.size && (t.size = a.size);
                    break;
                  default:
                    t = typeof a.is == "string" ? u.createElement(e, { is: a.is }) : u.createElement(e);
                }
            }
            t[el] = l, t[sl] = a;
            t: for (u = l.child; u !== null; ) {
              if (u.tag === 5 || u.tag === 6)
                t.appendChild(u.stateNode);
              else if (u.tag !== 4 && u.tag !== 27 && u.child !== null) {
                u.child.return = u, u = u.child;
                continue;
              }
              if (u === l) break t;
              for (; u.sibling === null; ) {
                if (u.return === null || u.return === l)
                  break t;
                u = u.return;
              }
              u.sibling.return = u.return, u = u.sibling;
            }
            l.stateNode = t;
            t: switch (tl(t, e, a), e) {
              case "button":
              case "input":
              case "select":
              case "textarea":
                t = !!a.autoFocus;
                break t;
              case "img":
                t = !0;
                break t;
              default:
                t = !1;
            }
            t && de(l);
          }
        }
        return Nt(l), l.flags &= -16777217, null;
      case 6:
        if (t && l.stateNode != null)
          t.memoizedProps !== a && de(l);
        else {
          if (typeof a != "string" && l.stateNode === null)
            throw Error(o(166));
          if (t = $.current, ou(l)) {
            if (t = l.stateNode, e = l.memoizedProps, a = null, u = fl, u !== null)
              switch (u.tag) {
                case 27:
                case 5:
                  a = u.memoizedProps;
              }
            t[el] = l, t = !!(t.nodeValue === e || a !== null && a.suppressHydrationWarning === !0 || Wo(t.nodeValue, e)), t || la(l);
          } else
            t = Fn(t).createTextNode(
              a
            ), t[el] = l, l.stateNode = t;
        }
        return Nt(l), null;
      case 13:
        if (a = l.memoizedState, t === null || t.memoizedState !== null && t.memoizedState.dehydrated !== null) {
          if (u = ou(l), a !== null && a.dehydrated !== null) {
            if (t === null) {
              if (!u) throw Error(o(318));
              if (u = l.memoizedState, u = u !== null ? u.dehydrated : null, !u) throw Error(o(317));
              u[el] = l;
            } else
              du(), (l.flags & 128) === 0 && (l.memoizedState = null), l.flags |= 4;
            Nt(l), u = !1;
          } else
            u = Cs(), t !== null && t.memoizedState !== null && (t.memoizedState.hydrationErrors = u), u = !0;
          if (!u)
            return l.flags & 256 ? (re(l), l) : (re(l), null);
        }
        if (re(l), (l.flags & 128) !== 0)
          return l.lanes = e, l;
        if (e = a !== null, t = t !== null && t.memoizedState !== null, e) {
          a = l.child, u = null, a.alternate !== null && a.alternate.memoizedState !== null && a.alternate.memoizedState.cachePool !== null && (u = a.alternate.memoizedState.cachePool.pool);
          var n = null;
          a.memoizedState !== null && a.memoizedState.cachePool !== null && (n = a.memoizedState.cachePool.pool), n !== u && (a.flags |= 2048);
        }
        return e !== t && e && (l.child.flags |= 8192), Yn(l, l.updateQueue), Nt(l), null;
      case 4:
        return Zl(), t === null && cf(l.stateNode.containerInfo), Nt(l), null;
      case 10:
        return fe(l.type), Nt(l), null;
      case 19:
        if (Y(Xt), u = l.memoizedState, u === null) return Nt(l), null;
        if (a = (l.flags & 128) !== 0, n = u.rendering, n === null)
          if (a) Nu(u, !1);
          else {
            if (Rt !== 0 || t !== null && (t.flags & 128) !== 0)
              for (t = l.child; t !== null; ) {
                if (n = Un(t), n !== null) {
                  for (l.flags |= 128, Nu(u, !1), t = n.updateQueue, l.updateQueue = t, Yn(l, t), l.subtreeFlags = 0, t = e, e = l.child; e !== null; )
                    Os(e, t), e = e.sibling;
                  return q(
                    Xt,
                    Xt.current & 1 | 2
                  ), l.child;
                }
                t = t.sibling;
              }
            u.tail !== null && gl() > Qn && (l.flags |= 128, a = !0, Nu(u, !1), l.lanes = 4194304);
          }
        else {
          if (!a)
            if (t = Un(n), t !== null) {
              if (l.flags |= 128, a = !0, t = t.updateQueue, l.updateQueue = t, Yn(l, t), Nu(u, !0), u.tail === null && u.tailMode === "hidden" && !n.alternate && !yt)
                return Nt(l), null;
            } else
              2 * gl() - u.renderingStartTime > Qn && e !== 536870912 && (l.flags |= 128, a = !0, Nu(u, !1), l.lanes = 4194304);
          u.isBackwards ? (n.sibling = l.child, l.child = n) : (t = u.last, t !== null ? t.sibling = n : l.child = n, u.last = n);
        }
        return u.tail !== null ? (l = u.tail, u.rendering = l, u.tail = l.sibling, u.renderingStartTime = gl(), l.sibling = null, t = Xt.current, q(Xt, a ? t & 1 | 2 : t & 1), l) : (Nt(l), null);
      case 22:
      case 23:
        return re(l), fc(), a = l.memoizedState !== null, t !== null ? t.memoizedState !== null !== a && (l.flags |= 8192) : a && (l.flags |= 8192), a ? (e & 536870912) !== 0 && (l.flags & 128) === 0 && (Nt(l), l.subtreeFlags & 6 && (l.flags |= 8192)) : Nt(l), e = l.updateQueue, e !== null && Yn(l, e.retryQueue), e = null, t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), a = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (a = l.memoizedState.cachePool.pool), a !== e && (l.flags |= 2048), t !== null && Y(ua), null;
      case 24:
        return e = null, t !== null && (e = t.memoizedState.cache), l.memoizedState.cache !== e && (l.flags |= 2048), fe(Gt), Nt(l), null;
      case 25:
        return null;
      case 30:
        return null;
    }
    throw Error(o(156, l.tag));
  }
  function l0(t, l) {
    switch (ki(l), l.tag) {
      case 1:
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 3:
        return fe(Gt), Zl(), t = l.flags, (t & 65536) !== 0 && (t & 128) === 0 ? (l.flags = t & -65537 | 128, l) : null;
      case 26:
      case 27:
      case 5:
        return Se(l), null;
      case 13:
        if (re(l), t = l.memoizedState, t !== null && t.dehydrated !== null) {
          if (l.alternate === null)
            throw Error(o(340));
          du();
        }
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 19:
        return Y(Xt), null;
      case 4:
        return Zl(), null;
      case 10:
        return fe(l.type), null;
      case 22:
      case 23:
        return re(l), fc(), t !== null && Y(ua), t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 24:
        return fe(Gt), null;
      case 25:
        return null;
      default:
        return null;
    }
  }
  function ao(t, l) {
    switch (ki(l), l.tag) {
      case 3:
        fe(Gt), Zl();
        break;
      case 26:
      case 27:
      case 5:
        Se(l);
        break;
      case 4:
        Zl();
        break;
      case 13:
        re(l);
        break;
      case 19:
        Y(Xt);
        break;
      case 10:
        fe(l.type);
        break;
      case 22:
      case 23:
        re(l), fc(), t !== null && Y(ua);
        break;
      case 24:
        fe(Gt);
    }
  }
  function Du(t, l) {
    try {
      var e = l.updateQueue, a = e !== null ? e.lastEffect : null;
      if (a !== null) {
        var u = a.next;
        e = u;
        do {
          if ((e.tag & t) === t) {
            a = void 0;
            var n = e.create, i = e.inst;
            a = n(), i.destroy = a;
          }
          e = e.next;
        } while (e !== u);
      }
    } catch (f) {
      pt(l, l.return, f);
    }
  }
  function Ne(t, l, e) {
    try {
      var a = l.updateQueue, u = a !== null ? a.lastEffect : null;
      if (u !== null) {
        var n = u.next;
        a = n;
        do {
          if ((a.tag & t) === t) {
            var i = a.inst, f = i.destroy;
            if (f !== void 0) {
              i.destroy = void 0, u = l;
              var s = e, m = f;
              try {
                m();
              } catch (j) {
                pt(
                  u,
                  s,
                  j
                );
              }
            }
          }
          a = a.next;
        } while (a !== n);
      }
    } catch (j) {
      pt(l, l.return, j);
    }
  }
  function uo(t) {
    var l = t.updateQueue;
    if (l !== null) {
      var e = t.stateNode;
      try {
        ws(l, e);
      } catch (a) {
        pt(t, t.return, a);
      }
    }
  }
  function no(t, l, e) {
    e.props = ia(
      t.type,
      t.memoizedProps
    ), e.state = t.memoizedState;
    try {
      e.componentWillUnmount();
    } catch (a) {
      pt(t, l, a);
    }
  }
  function Ou(t, l) {
    try {
      var e = t.ref;
      if (e !== null) {
        switch (t.tag) {
          case 26:
          case 27:
          case 5:
            var a = t.stateNode;
            break;
          case 30:
            a = t.stateNode;
            break;
          default:
            a = t.stateNode;
        }
        typeof e == "function" ? t.refCleanup = e(a) : e.current = a;
      }
    } catch (u) {
      pt(t, l, u);
    }
  }
  function kl(t, l) {
    var e = t.ref, a = t.refCleanup;
    if (e !== null)
      if (typeof a == "function")
        try {
          a();
        } catch (u) {
          pt(t, l, u);
        } finally {
          t.refCleanup = null, t = t.alternate, t != null && (t.refCleanup = null);
        }
      else if (typeof e == "function")
        try {
          e(null);
        } catch (u) {
          pt(t, l, u);
        }
      else e.current = null;
  }
  function io(t) {
    var l = t.type, e = t.memoizedProps, a = t.stateNode;
    try {
      t: switch (l) {
        case "button":
        case "input":
        case "select":
        case "textarea":
          e.autoFocus && a.focus();
          break t;
        case "img":
          e.src ? a.src = e.src : e.srcSet && (a.srcset = e.srcSet);
      }
    } catch (u) {
      pt(t, t.return, u);
    }
  }
  function qc(t, l, e) {
    try {
      var a = t.stateNode;
      j0(a, t.type, e, l), a[sl] = l;
    } catch (u) {
      pt(t, t.return, u);
    }
  }
  function co(t) {
    return t.tag === 5 || t.tag === 3 || t.tag === 26 || t.tag === 27 && qe(t.type) || t.tag === 4;
  }
  function Bc(t) {
    t: for (; ; ) {
      for (; t.sibling === null; ) {
        if (t.return === null || co(t.return)) return null;
        t = t.return;
      }
      for (t.sibling.return = t.return, t = t.sibling; t.tag !== 5 && t.tag !== 6 && t.tag !== 18; ) {
        if (t.tag === 27 && qe(t.type) || t.flags & 2 || t.child === null || t.tag === 4) continue t;
        t.child.return = t, t = t.child;
      }
      if (!(t.flags & 2)) return t.stateNode;
    }
  }
  function Yc(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? (e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e).insertBefore(t, l) : (l = e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, l.appendChild(t), e = e._reactRootContainer, e != null || l.onclick !== null || (l.onclick = Wn));
    else if (a !== 4 && (a === 27 && qe(t.type) && (e = t.stateNode, l = null), t = t.child, t !== null))
      for (Yc(t, l, e), t = t.sibling; t !== null; )
        Yc(t, l, e), t = t.sibling;
  }
  function Gn(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? e.insertBefore(t, l) : e.appendChild(t);
    else if (a !== 4 && (a === 27 && qe(t.type) && (e = t.stateNode), t = t.child, t !== null))
      for (Gn(t, l, e), t = t.sibling; t !== null; )
        Gn(t, l, e), t = t.sibling;
  }
  function fo(t) {
    var l = t.stateNode, e = t.memoizedProps;
    try {
      for (var a = t.type, u = l.attributes; u.length; )
        l.removeAttributeNode(u[0]);
      tl(l, a, e), l[el] = t, l[sl] = e;
    } catch (n) {
      pt(t, t.return, n);
    }
  }
  var he = !1, Ct = !1, Gc = !1, so = typeof WeakSet == "function" ? WeakSet : Set, Jt = null;
  function e0(t, l) {
    if (t = t.containerInfo, rf = ai, t = xs(t), Bi(t)) {
      if ("selectionStart" in t)
        var e = {
          start: t.selectionStart,
          end: t.selectionEnd
        };
      else
        t: {
          e = (e = t.ownerDocument) && e.defaultView || window;
          var a = e.getSelection && e.getSelection();
          if (a && a.rangeCount !== 0) {
            e = a.anchorNode;
            var u = a.anchorOffset, n = a.focusNode;
            a = a.focusOffset;
            try {
              e.nodeType, n.nodeType;
            } catch {
              e = null;
              break t;
            }
            var i = 0, f = -1, s = -1, m = 0, j = 0, z = t, g = null;
            l: for (; ; ) {
              for (var S; z !== e || u !== 0 && z.nodeType !== 3 || (f = i + u), z !== n || a !== 0 && z.nodeType !== 3 || (s = i + a), z.nodeType === 3 && (i += z.nodeValue.length), (S = z.firstChild) !== null; )
                g = z, z = S;
              for (; ; ) {
                if (z === t) break l;
                if (g === e && ++m === u && (f = i), g === n && ++j === a && (s = i), (S = z.nextSibling) !== null) break;
                z = g, g = z.parentNode;
              }
              z = S;
            }
            e = f === -1 || s === -1 ? null : { start: f, end: s };
          } else e = null;
        }
      e = e || { start: 0, end: 0 };
    } else e = null;
    for (of = { focusedElem: t, selectionRange: e }, ai = !1, Jt = l; Jt !== null; )
      if (l = Jt, t = l.child, (l.subtreeFlags & 1024) !== 0 && t !== null)
        t.return = l, Jt = t;
      else
        for (; Jt !== null; ) {
          switch (l = Jt, n = l.alternate, t = l.flags, l.tag) {
            case 0:
              break;
            case 11:
            case 15:
              break;
            case 1:
              if ((t & 1024) !== 0 && n !== null) {
                t = void 0, e = l, u = n.memoizedProps, n = n.memoizedState, a = e.stateNode;
                try {
                  var k = ia(
                    e.type,
                    u,
                    e.elementType === e.type
                  );
                  t = a.getSnapshotBeforeUpdate(
                    k,
                    n
                  ), a.__reactInternalSnapshotBeforeUpdate = t;
                } catch (K) {
                  pt(
                    e,
                    e.return,
                    K
                  );
                }
              }
              break;
            case 3:
              if ((t & 1024) !== 0) {
                if (t = l.stateNode.containerInfo, e = t.nodeType, e === 9)
                  yf(t);
                else if (e === 1)
                  switch (t.nodeName) {
                    case "HEAD":
                    case "HTML":
                    case "BODY":
                      yf(t);
                      break;
                    default:
                      t.textContent = "";
                  }
              }
              break;
            case 5:
            case 26:
            case 27:
            case 6:
            case 4:
            case 17:
              break;
            default:
              if ((t & 1024) !== 0) throw Error(o(163));
          }
          if (t = l.sibling, t !== null) {
            t.return = l.return, Jt = t;
            break;
          }
          Jt = l.return;
        }
  }
  function ro(t, l, e) {
    var a = e.flags;
    switch (e.tag) {
      case 0:
      case 11:
      case 15:
        De(t, e), a & 4 && Du(5, e);
        break;
      case 1:
        if (De(t, e), a & 4)
          if (t = e.stateNode, l === null)
            try {
              t.componentDidMount();
            } catch (i) {
              pt(e, e.return, i);
            }
          else {
            var u = ia(
              e.type,
              l.memoizedProps
            );
            l = l.memoizedState;
            try {
              t.componentDidUpdate(
                u,
                l,
                t.__reactInternalSnapshotBeforeUpdate
              );
            } catch (i) {
              pt(
                e,
                e.return,
                i
              );
            }
          }
        a & 64 && uo(e), a & 512 && Ou(e, e.return);
        break;
      case 3:
        if (De(t, e), a & 64 && (t = e.updateQueue, t !== null)) {
          if (l = null, e.child !== null)
            switch (e.child.tag) {
              case 27:
              case 5:
                l = e.child.stateNode;
                break;
              case 1:
                l = e.child.stateNode;
            }
          try {
            ws(t, l);
          } catch (i) {
            pt(e, e.return, i);
          }
        }
        break;
      case 27:
        l === null && a & 4 && fo(e);
      case 26:
      case 5:
        De(t, e), l === null && a & 4 && io(e), a & 512 && Ou(e, e.return);
        break;
      case 12:
        De(t, e);
        break;
      case 13:
        De(t, e), a & 4 && yo(t, e), a & 64 && (t = e.memoizedState, t !== null && (t = t.dehydrated, t !== null && (e = o0.bind(
          null,
          e
        ), D0(t, e))));
        break;
      case 22:
        if (a = e.memoizedState !== null || he, !a) {
          l = l !== null && l.memoizedState !== null || Ct, u = he;
          var n = Ct;
          he = a, (Ct = l) && !n ? Oe(
            t,
            e,
            (e.subtreeFlags & 8772) !== 0
          ) : De(t, e), he = u, Ct = n;
        }
        break;
      case 30:
        break;
      default:
        De(t, e);
    }
  }
  function oo(t) {
    var l = t.alternate;
    l !== null && (t.alternate = null, oo(l)), t.child = null, t.deletions = null, t.sibling = null, t.tag === 5 && (l = t.stateNode, l !== null && Si(l)), t.stateNode = null, t.return = null, t.dependencies = null, t.memoizedProps = null, t.memoizedState = null, t.pendingProps = null, t.stateNode = null, t.updateQueue = null;
  }
  var Mt = null, dl = !1;
  function ye(t, l, e) {
    for (e = e.child; e !== null; )
      ho(t, l, e), e = e.sibling;
  }
  function ho(t, l, e) {
    if (Ft && typeof Ft.onCommitFiberUnmount == "function")
      try {
        Ft.onCommitFiberUnmount(Lt, e);
      } catch {
      }
    switch (e.tag) {
      case 26:
        Ct || kl(e, l), ye(
          t,
          l,
          e
        ), e.memoizedState ? e.memoizedState.count-- : e.stateNode && (e = e.stateNode, e.parentNode.removeChild(e));
        break;
      case 27:
        Ct || kl(e, l);
        var a = Mt, u = dl;
        qe(e.type) && (Mt = e.stateNode, dl = !1), ye(
          t,
          l,
          e
        ), Gu(e.stateNode), Mt = a, dl = u;
        break;
      case 5:
        Ct || kl(e, l);
      case 6:
        if (a = Mt, u = dl, Mt = null, ye(
          t,
          l,
          e
        ), Mt = a, dl = u, Mt !== null)
          if (dl)
            try {
              (Mt.nodeType === 9 ? Mt.body : Mt.nodeName === "HTML" ? Mt.ownerDocument.body : Mt).removeChild(e.stateNode);
            } catch (n) {
              pt(
                e,
                l,
                n
              );
            }
          else
            try {
              Mt.removeChild(e.stateNode);
            } catch (n) {
              pt(
                e,
                l,
                n
              );
            }
        break;
      case 18:
        Mt !== null && (dl ? (t = Mt, ld(
          t.nodeType === 9 ? t.body : t.nodeName === "HTML" ? t.ownerDocument.body : t,
          e.stateNode
        ), Ju(t)) : ld(Mt, e.stateNode));
        break;
      case 4:
        a = Mt, u = dl, Mt = e.stateNode.containerInfo, dl = !0, ye(
          t,
          l,
          e
        ), Mt = a, dl = u;
        break;
      case 0:
      case 11:
      case 14:
      case 15:
        Ct || Ne(2, e, l), Ct || Ne(4, e, l), ye(
          t,
          l,
          e
        );
        break;
      case 1:
        Ct || (kl(e, l), a = e.stateNode, typeof a.componentWillUnmount == "function" && no(
          e,
          l,
          a
        )), ye(
          t,
          l,
          e
        );
        break;
      case 21:
        ye(
          t,
          l,
          e
        );
        break;
      case 22:
        Ct = (a = Ct) || e.memoizedState !== null, ye(
          t,
          l,
          e
        ), Ct = a;
        break;
      default:
        ye(
          t,
          l,
          e
        );
    }
  }
  function yo(t, l) {
    if (l.memoizedState === null && (t = l.alternate, t !== null && (t = t.memoizedState, t !== null && (t = t.dehydrated, t !== null))))
      try {
        Ju(t);
      } catch (e) {
        pt(l, l.return, e);
      }
  }
  function a0(t) {
    switch (t.tag) {
      case 13:
      case 19:
        var l = t.stateNode;
        return l === null && (l = t.stateNode = new so()), l;
      case 22:
        return t = t.stateNode, l = t._retryCache, l === null && (l = t._retryCache = new so()), l;
      default:
        throw Error(o(435, t.tag));
    }
  }
  function Xc(t, l) {
    var e = a0(t);
    l.forEach(function(a) {
      var u = d0.bind(null, t, a);
      e.has(a) || (e.add(a), a.then(u, u));
    });
  }
  function jl(t, l) {
    var e = l.deletions;
    if (e !== null)
      for (var a = 0; a < e.length; a++) {
        var u = e[a], n = t, i = l, f = i;
        t: for (; f !== null; ) {
          switch (f.tag) {
            case 27:
              if (qe(f.type)) {
                Mt = f.stateNode, dl = !1;
                break t;
              }
              break;
            case 5:
              Mt = f.stateNode, dl = !1;
              break t;
            case 3:
            case 4:
              Mt = f.stateNode.containerInfo, dl = !0;
              break t;
          }
          f = f.return;
        }
        if (Mt === null) throw Error(o(160));
        ho(n, i, u), Mt = null, dl = !1, n = u.alternate, n !== null && (n.return = null), u.return = null;
      }
    if (l.subtreeFlags & 13878)
      for (l = l.child; l !== null; )
        vo(l, t), l = l.sibling;
  }
  var Gl = null;
  function vo(t, l) {
    var e = t.alternate, a = t.flags;
    switch (t.tag) {
      case 0:
      case 11:
      case 14:
      case 15:
        jl(l, t), Tl(t), a & 4 && (Ne(3, t, t.return), Du(3, t), Ne(5, t, t.return));
        break;
      case 1:
        jl(l, t), Tl(t), a & 512 && (Ct || e === null || kl(e, e.return)), a & 64 && he && (t = t.updateQueue, t !== null && (a = t.callbacks, a !== null && (e = t.shared.hiddenCallbacks, t.shared.hiddenCallbacks = e === null ? a : e.concat(a))));
        break;
      case 26:
        var u = Gl;
        if (jl(l, t), Tl(t), a & 512 && (Ct || e === null || kl(e, e.return)), a & 4) {
          var n = e !== null ? e.memoizedState : null;
          if (a = t.memoizedState, e === null)
            if (a === null)
              if (t.stateNode === null) {
                t: {
                  a = t.type, e = t.memoizedProps, u = u.ownerDocument || u;
                  l: switch (a) {
                    case "title":
                      n = u.getElementsByTagName("title")[0], (!n || n[tu] || n[el] || n.namespaceURI === "http://www.w3.org/2000/svg" || n.hasAttribute("itemprop")) && (n = u.createElement(a), u.head.insertBefore(
                        n,
                        u.querySelector("head > title")
                      )), tl(n, a, e), n[el] = t, wt(n), a = n;
                      break t;
                    case "link":
                      var i = sd(
                        "link",
                        "href",
                        u
                      ).get(a + (e.href || ""));
                      if (i) {
                        for (var f = 0; f < i.length; f++)
                          if (n = i[f], n.getAttribute("href") === (e.href == null || e.href === "" ? null : e.href) && n.getAttribute("rel") === (e.rel == null ? null : e.rel) && n.getAttribute("title") === (e.title == null ? null : e.title) && n.getAttribute("crossorigin") === (e.crossOrigin == null ? null : e.crossOrigin)) {
                            i.splice(f, 1);
                            break l;
                          }
                      }
                      n = u.createElement(a), tl(n, a, e), u.head.appendChild(n);
                      break;
                    case "meta":
                      if (i = sd(
                        "meta",
                        "content",
                        u
                      ).get(a + (e.content || ""))) {
                        for (f = 0; f < i.length; f++)
                          if (n = i[f], n.getAttribute("content") === (e.content == null ? null : "" + e.content) && n.getAttribute("name") === (e.name == null ? null : e.name) && n.getAttribute("property") === (e.property == null ? null : e.property) && n.getAttribute("http-equiv") === (e.httpEquiv == null ? null : e.httpEquiv) && n.getAttribute("charset") === (e.charSet == null ? null : e.charSet)) {
                            i.splice(f, 1);
                            break l;
                          }
                      }
                      n = u.createElement(a), tl(n, a, e), u.head.appendChild(n);
                      break;
                    default:
                      throw Error(o(468, a));
                  }
                  n[el] = t, wt(n), a = n;
                }
                t.stateNode = a;
              } else
                rd(
                  u,
                  t.type,
                  t.stateNode
                );
            else
              t.stateNode = fd(
                u,
                a,
                t.memoizedProps
              );
          else
            n !== a ? (n === null ? e.stateNode !== null && (e = e.stateNode, e.parentNode.removeChild(e)) : n.count--, a === null ? rd(
              u,
              t.type,
              t.stateNode
            ) : fd(
              u,
              a,
              t.memoizedProps
            )) : a === null && t.stateNode !== null && qc(
              t,
              t.memoizedProps,
              e.memoizedProps
            );
        }
        break;
      case 27:
        jl(l, t), Tl(t), a & 512 && (Ct || e === null || kl(e, e.return)), e !== null && a & 4 && qc(
          t,
          t.memoizedProps,
          e.memoizedProps
        );
        break;
      case 5:
        if (jl(l, t), Tl(t), a & 512 && (Ct || e === null || kl(e, e.return)), t.flags & 32) {
          u = t.stateNode;
          try {
            xa(u, "");
          } catch (S) {
            pt(t, t.return, S);
          }
        }
        a & 4 && t.stateNode != null && (u = t.memoizedProps, qc(
          t,
          u,
          e !== null ? e.memoizedProps : u
        )), a & 1024 && (Gc = !0);
        break;
      case 6:
        if (jl(l, t), Tl(t), a & 4) {
          if (t.stateNode === null)
            throw Error(o(162));
          a = t.memoizedProps, e = t.stateNode;
          try {
            e.nodeValue = a;
          } catch (S) {
            pt(t, t.return, S);
          }
        }
        break;
      case 3:
        if (ti = null, u = Gl, Gl = Pn(l.containerInfo), jl(l, t), Gl = u, Tl(t), a & 4 && e !== null && e.memoizedState.isDehydrated)
          try {
            Ju(l.containerInfo);
          } catch (S) {
            pt(t, t.return, S);
          }
        Gc && (Gc = !1, mo(t));
        break;
      case 4:
        a = Gl, Gl = Pn(
          t.stateNode.containerInfo
        ), jl(l, t), Tl(t), Gl = a;
        break;
      case 12:
        jl(l, t), Tl(t);
        break;
      case 13:
        jl(l, t), Tl(t), t.child.flags & 8192 && t.memoizedState !== null != (e !== null && e.memoizedState !== null) && (Kc = gl()), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Xc(t, a)));
        break;
      case 22:
        u = t.memoizedState !== null;
        var s = e !== null && e.memoizedState !== null, m = he, j = Ct;
        if (he = m || u, Ct = j || s, jl(l, t), Ct = j, he = m, Tl(t), a & 8192)
          t: for (l = t.stateNode, l._visibility = u ? l._visibility & -2 : l._visibility | 1, u && (e === null || s || he || Ct || ca(t)), e = null, l = t; ; ) {
            if (l.tag === 5 || l.tag === 26) {
              if (e === null) {
                s = e = l;
                try {
                  if (n = s.stateNode, u)
                    i = n.style, typeof i.setProperty == "function" ? i.setProperty("display", "none", "important") : i.display = "none";
                  else {
                    f = s.stateNode;
                    var z = s.memoizedProps.style, g = z != null && z.hasOwnProperty("display") ? z.display : null;
                    f.style.display = g == null || typeof g == "boolean" ? "" : ("" + g).trim();
                  }
                } catch (S) {
                  pt(s, s.return, S);
                }
              }
            } else if (l.tag === 6) {
              if (e === null) {
                s = l;
                try {
                  s.stateNode.nodeValue = u ? "" : s.memoizedProps;
                } catch (S) {
                  pt(s, s.return, S);
                }
              }
            } else if ((l.tag !== 22 && l.tag !== 23 || l.memoizedState === null || l === t) && l.child !== null) {
              l.child.return = l, l = l.child;
              continue;
            }
            if (l === t) break t;
            for (; l.sibling === null; ) {
              if (l.return === null || l.return === t) break t;
              e === l && (e = null), l = l.return;
            }
            e === l && (e = null), l.sibling.return = l.return, l = l.sibling;
          }
        a & 4 && (a = t.updateQueue, a !== null && (e = a.retryQueue, e !== null && (a.retryQueue = null, Xc(t, e))));
        break;
      case 19:
        jl(l, t), Tl(t), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Xc(t, a)));
        break;
      case 30:
        break;
      case 21:
        break;
      default:
        jl(l, t), Tl(t);
    }
  }
  function Tl(t) {
    var l = t.flags;
    if (l & 2) {
      try {
        for (var e, a = t.return; a !== null; ) {
          if (co(a)) {
            e = a;
            break;
          }
          a = a.return;
        }
        if (e == null) throw Error(o(160));
        switch (e.tag) {
          case 27:
            var u = e.stateNode, n = Bc(t);
            Gn(t, n, u);
            break;
          case 5:
            var i = e.stateNode;
            e.flags & 32 && (xa(i, ""), e.flags &= -33);
            var f = Bc(t);
            Gn(t, f, i);
            break;
          case 3:
          case 4:
            var s = e.stateNode.containerInfo, m = Bc(t);
            Yc(
              t,
              m,
              s
            );
            break;
          default:
            throw Error(o(161));
        }
      } catch (j) {
        pt(t, t.return, j);
      }
      t.flags &= -3;
    }
    l & 4096 && (t.flags &= -4097);
  }
  function mo(t) {
    if (t.subtreeFlags & 1024)
      for (t = t.child; t !== null; ) {
        var l = t;
        mo(l), l.tag === 5 && l.flags & 1024 && l.stateNode.reset(), t = t.sibling;
      }
  }
  function De(t, l) {
    if (l.subtreeFlags & 8772)
      for (l = l.child; l !== null; )
        ro(t, l.alternate, l), l = l.sibling;
  }
  function ca(t) {
    for (t = t.child; t !== null; ) {
      var l = t;
      switch (l.tag) {
        case 0:
        case 11:
        case 14:
        case 15:
          Ne(4, l, l.return), ca(l);
          break;
        case 1:
          kl(l, l.return);
          var e = l.stateNode;
          typeof e.componentWillUnmount == "function" && no(
            l,
            l.return,
            e
          ), ca(l);
          break;
        case 27:
          Gu(l.stateNode);
        case 26:
        case 5:
          kl(l, l.return), ca(l);
          break;
        case 22:
          l.memoizedState === null && ca(l);
          break;
        case 30:
          ca(l);
          break;
        default:
          ca(l);
      }
      t = t.sibling;
    }
  }
  function Oe(t, l, e) {
    for (e = e && (l.subtreeFlags & 8772) !== 0, l = l.child; l !== null; ) {
      var a = l.alternate, u = t, n = l, i = n.flags;
      switch (n.tag) {
        case 0:
        case 11:
        case 15:
          Oe(
            u,
            n,
            e
          ), Du(4, n);
          break;
        case 1:
          if (Oe(
            u,
            n,
            e
          ), a = n, u = a.stateNode, typeof u.componentDidMount == "function")
            try {
              u.componentDidMount();
            } catch (m) {
              pt(a, a.return, m);
            }
          if (a = n, u = a.updateQueue, u !== null) {
            var f = a.stateNode;
            try {
              var s = u.shared.hiddenCallbacks;
              if (s !== null)
                for (u.shared.hiddenCallbacks = null, u = 0; u < s.length; u++)
                  Vs(s[u], f);
            } catch (m) {
              pt(a, a.return, m);
            }
          }
          e && i & 64 && uo(n), Ou(n, n.return);
          break;
        case 27:
          fo(n);
        case 26:
        case 5:
          Oe(
            u,
            n,
            e
          ), e && a === null && i & 4 && io(n), Ou(n, n.return);
          break;
        case 12:
          Oe(
            u,
            n,
            e
          );
          break;
        case 13:
          Oe(
            u,
            n,
            e
          ), e && i & 4 && yo(u, n);
          break;
        case 22:
          n.memoizedState === null && Oe(
            u,
            n,
            e
          ), Ou(n, n.return);
          break;
        case 30:
          break;
        default:
          Oe(
            u,
            n,
            e
          );
      }
      l = l.sibling;
    }
  }
  function Qc(t, l) {
    var e = null;
    t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), t = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (t = l.memoizedState.cachePool.pool), t !== e && (t != null && t.refCount++, e != null && vu(e));
  }
  function Zc(t, l) {
    t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && vu(t));
  }
  function $l(t, l, e, a) {
    if (l.subtreeFlags & 10256)
      for (l = l.child; l !== null; )
        go(
          t,
          l,
          e,
          a
        ), l = l.sibling;
  }
  function go(t, l, e, a) {
    var u = l.flags;
    switch (l.tag) {
      case 0:
      case 11:
      case 15:
        $l(
          t,
          l,
          e,
          a
        ), u & 2048 && Du(9, l);
        break;
      case 1:
        $l(
          t,
          l,
          e,
          a
        );
        break;
      case 3:
        $l(
          t,
          l,
          e,
          a
        ), u & 2048 && (t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && vu(t)));
        break;
      case 12:
        if (u & 2048) {
          $l(
            t,
            l,
            e,
            a
          ), t = l.stateNode;
          try {
            var n = l.memoizedProps, i = n.id, f = n.onPostCommit;
            typeof f == "function" && f(
              i,
              l.alternate === null ? "mount" : "update",
              t.passiveEffectDuration,
              -0
            );
          } catch (s) {
            pt(l, l.return, s);
          }
        } else
          $l(
            t,
            l,
            e,
            a
          );
        break;
      case 13:
        $l(
          t,
          l,
          e,
          a
        );
        break;
      case 23:
        break;
      case 22:
        n = l.stateNode, i = l.alternate, l.memoizedState !== null ? n._visibility & 2 ? $l(
          t,
          l,
          e,
          a
        ) : Ru(t, l) : n._visibility & 2 ? $l(
          t,
          l,
          e,
          a
        ) : (n._visibility |= 2, Ga(
          t,
          l,
          e,
          a,
          (l.subtreeFlags & 10256) !== 0
        )), u & 2048 && Qc(i, l);
        break;
      case 24:
        $l(
          t,
          l,
          e,
          a
        ), u & 2048 && Zc(l.alternate, l);
        break;
      default:
        $l(
          t,
          l,
          e,
          a
        );
    }
  }
  function Ga(t, l, e, a, u) {
    for (u = u && (l.subtreeFlags & 10256) !== 0, l = l.child; l !== null; ) {
      var n = t, i = l, f = e, s = a, m = i.flags;
      switch (i.tag) {
        case 0:
        case 11:
        case 15:
          Ga(
            n,
            i,
            f,
            s,
            u
          ), Du(8, i);
          break;
        case 23:
          break;
        case 22:
          var j = i.stateNode;
          i.memoizedState !== null ? j._visibility & 2 ? Ga(
            n,
            i,
            f,
            s,
            u
          ) : Ru(
            n,
            i
          ) : (j._visibility |= 2, Ga(
            n,
            i,
            f,
            s,
            u
          )), u && m & 2048 && Qc(
            i.alternate,
            i
          );
          break;
        case 24:
          Ga(
            n,
            i,
            f,
            s,
            u
          ), u && m & 2048 && Zc(i.alternate, i);
          break;
        default:
          Ga(
            n,
            i,
            f,
            s,
            u
          );
      }
      l = l.sibling;
    }
  }
  function Ru(t, l) {
    if (l.subtreeFlags & 10256)
      for (l = l.child; l !== null; ) {
        var e = t, a = l, u = a.flags;
        switch (a.tag) {
          case 22:
            Ru(e, a), u & 2048 && Qc(
              a.alternate,
              a
            );
            break;
          case 24:
            Ru(e, a), u & 2048 && Zc(a.alternate, a);
            break;
          default:
            Ru(e, a);
        }
        l = l.sibling;
      }
  }
  var _u = 8192;
  function Xa(t) {
    if (t.subtreeFlags & _u)
      for (t = t.child; t !== null; )
        So(t), t = t.sibling;
  }
  function So(t) {
    switch (t.tag) {
      case 26:
        Xa(t), t.flags & _u && t.memoizedState !== null && Z0(
          Gl,
          t.memoizedState,
          t.memoizedProps
        );
        break;
      case 5:
        Xa(t);
        break;
      case 3:
      case 4:
        var l = Gl;
        Gl = Pn(t.stateNode.containerInfo), Xa(t), Gl = l;
        break;
      case 22:
        t.memoizedState === null && (l = t.alternate, l !== null && l.memoizedState !== null ? (l = _u, _u = 16777216, Xa(t), _u = l) : Xa(t));
        break;
      default:
        Xa(t);
    }
  }
  function bo(t) {
    var l = t.alternate;
    if (l !== null && (t = l.child, t !== null)) {
      l.child = null;
      do
        l = t.sibling, t.sibling = null, t = l;
      while (t !== null);
    }
  }
  function Uu(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          Jt = a, po(
            a,
            t
          );
        }
      bo(t);
    }
    if (t.subtreeFlags & 10256)
      for (t = t.child; t !== null; )
        xo(t), t = t.sibling;
  }
  function xo(t) {
    switch (t.tag) {
      case 0:
      case 11:
      case 15:
        Uu(t), t.flags & 2048 && Ne(9, t, t.return);
        break;
      case 3:
        Uu(t);
        break;
      case 12:
        Uu(t);
        break;
      case 22:
        var l = t.stateNode;
        t.memoizedState !== null && l._visibility & 2 && (t.return === null || t.return.tag !== 13) ? (l._visibility &= -3, Xn(t)) : Uu(t);
        break;
      default:
        Uu(t);
    }
  }
  function Xn(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          Jt = a, po(
            a,
            t
          );
        }
      bo(t);
    }
    for (t = t.child; t !== null; ) {
      switch (l = t, l.tag) {
        case 0:
        case 11:
        case 15:
          Ne(8, l, l.return), Xn(l);
          break;
        case 22:
          e = l.stateNode, e._visibility & 2 && (e._visibility &= -3, Xn(l));
          break;
        default:
          Xn(l);
      }
      t = t.sibling;
    }
  }
  function po(t, l) {
    for (; Jt !== null; ) {
      var e = Jt;
      switch (e.tag) {
        case 0:
        case 11:
        case 15:
          Ne(8, e, l);
          break;
        case 23:
        case 22:
          if (e.memoizedState !== null && e.memoizedState.cachePool !== null) {
            var a = e.memoizedState.cachePool.pool;
            a != null && a.refCount++;
          }
          break;
        case 24:
          vu(e.memoizedState.cache);
      }
      if (a = e.child, a !== null) a.return = e, Jt = a;
      else
        t: for (e = t; Jt !== null; ) {
          a = Jt;
          var u = a.sibling, n = a.return;
          if (oo(a), a === e) {
            Jt = null;
            break t;
          }
          if (u !== null) {
            u.return = n, Jt = u;
            break t;
          }
          Jt = n;
        }
    }
  }
  var u0 = {
    getCacheForType: function(t) {
      var l = al(Gt), e = l.data.get(t);
      return e === void 0 && (e = t(), l.data.set(t, e)), e;
    }
  }, n0 = typeof WeakMap == "function" ? WeakMap : Map, vt = 0, At = null, at = null, ft = 0, mt = 0, Al = null, Re = !1, Qa = !1, Lc = !1, ve = 0, Rt = 0, _e = 0, fa = 0, Vc = 0, Hl = 0, Za = 0, Cu = null, hl = null, wc = !1, Kc = 0, Qn = 1 / 0, Zn = null, Ue = null, It = 0, Ce = null, La = null, Va = 0, Jc = 0, kc = null, jo = null, Hu = 0, $c = null;
  function El() {
    if ((vt & 2) !== 0 && ft !== 0)
      return ft & -ft;
    if (p.T !== null) {
      var t = Ra;
      return t !== 0 ? t : ef();
    }
    return le();
  }
  function To() {
    Hl === 0 && (Hl = (ft & 536870912) === 0 || yt ? E() : 536870912);
    var t = Cl.current;
    return t !== null && (t.flags |= 32), Hl;
  }
  function zl(t, l, e) {
    (t === At && (mt === 2 || mt === 9) || t.cancelPendingCommit !== null) && (wa(t, 0), He(
      t,
      ft,
      Hl,
      !1
    )), tt(t, e), ((vt & 2) === 0 || t !== At) && (t === At && ((vt & 2) === 0 && (fa |= e), Rt === 4 && He(
      t,
      ft,
      Hl,
      !1
    )), Wl(t));
  }
  function Ao(t, l, e) {
    if ((vt & 6) !== 0) throw Error(o(327));
    var a = !e && (l & 124) === 0 && (l & t.expiredLanes) === 0 || Vl(t, l), u = a ? f0(t, l) : Pc(t, l, !0), n = a;
    do {
      if (u === 0) {
        Qa && !a && He(t, l, 0, !1);
        break;
      } else {
        if (e = t.current.alternate, n && !i0(e)) {
          u = Pc(t, l, !1), n = !1;
          continue;
        }
        if (u === 2) {
          if (n = l, t.errorRecoveryDisabledLanes & n)
            var i = 0;
          else
            i = t.pendingLanes & -536870913, i = i !== 0 ? i : i & 536870912 ? 536870912 : 0;
          if (i !== 0) {
            l = i;
            t: {
              var f = t;
              u = Cu;
              var s = f.current.memoizedState.isDehydrated;
              if (s && (wa(f, i).flags |= 256), i = Pc(
                f,
                i,
                !1
              ), i !== 2) {
                if (Lc && !s) {
                  f.errorRecoveryDisabledLanes |= n, fa |= n, u = 4;
                  break t;
                }
                n = hl, hl = u, n !== null && (hl === null ? hl = n : hl.push.apply(
                  hl,
                  n
                ));
              }
              u = i;
            }
            if (n = !1, u !== 2) continue;
          }
        }
        if (u === 1) {
          wa(t, 0), He(t, l, 0, !0);
          break;
        }
        t: {
          switch (a = t, n = u, n) {
            case 0:
            case 1:
              throw Error(o(345));
            case 4:
              if ((l & 4194048) !== l) break;
            case 6:
              He(
                a,
                l,
                Hl,
                !Re
              );
              break t;
            case 2:
              hl = null;
              break;
            case 3:
            case 5:
              break;
            default:
              throw Error(o(329));
          }
          if ((l & 62914560) === l && (u = Kc + 300 - gl(), 10 < u)) {
            if (He(
              a,
              l,
              Hl,
              !Re
            ), da(a, 0, !0) !== 0) break t;
            a.timeoutHandle = Io(
              Eo.bind(
                null,
                a,
                e,
                hl,
                Zn,
                wc,
                l,
                Hl,
                fa,
                Za,
                Re,
                n,
                2,
                -0,
                0
              ),
              u
            );
            break t;
          }
          Eo(
            a,
            e,
            hl,
            Zn,
            wc,
            l,
            Hl,
            fa,
            Za,
            Re,
            n,
            0,
            -0,
            0
          );
        }
      }
      break;
    } while (!0);
    Wl(t);
  }
  function Eo(t, l, e, a, u, n, i, f, s, m, j, z, g, S) {
    if (t.timeoutHandle = -1, z = l.subtreeFlags, (z & 8192 || (z & 16785408) === 16785408) && (Zu = { stylesheets: null, count: 0, unsuspend: Q0 }, So(l), z = L0(), z !== null)) {
      t.cancelPendingCommit = z(
        _o.bind(
          null,
          t,
          l,
          n,
          e,
          a,
          u,
          i,
          f,
          s,
          j,
          1,
          g,
          S
        )
      ), He(t, n, i, !m);
      return;
    }
    _o(
      t,
      l,
      n,
      e,
      a,
      u,
      i,
      f,
      s
    );
  }
  function i0(t) {
    for (var l = t; ; ) {
      var e = l.tag;
      if ((e === 0 || e === 11 || e === 15) && l.flags & 16384 && (e = l.updateQueue, e !== null && (e = e.stores, e !== null)))
        for (var a = 0; a < e.length; a++) {
          var u = e[a], n = u.getSnapshot;
          u = u.value;
          try {
            if (!xl(n(), u)) return !1;
          } catch {
            return !1;
          }
        }
      if (e = l.child, l.subtreeFlags & 16384 && e !== null)
        e.return = l, l = e;
      else {
        if (l === t) break;
        for (; l.sibling === null; ) {
          if (l.return === null || l.return === t) return !0;
          l = l.return;
        }
        l.sibling.return = l.return, l = l.sibling;
      }
    }
    return !0;
  }
  function He(t, l, e, a) {
    l &= ~Vc, l &= ~fa, t.suspendedLanes |= l, t.pingedLanes &= ~l, a && (t.warmLanes |= l), a = t.expirationTimes;
    for (var u = l; 0 < u; ) {
      var n = 31 - ct(u), i = 1 << n;
      a[n] = -1, u &= ~i;
    }
    e !== 0 && Vt(t, e, l);
  }
  function Ln() {
    return (vt & 6) === 0 ? (qu(0), !1) : !0;
  }
  function Wc() {
    if (at !== null) {
      if (mt === 0)
        var t = at.return;
      else
        t = at, ce = ea = null, hc(t), Ba = null, zu = 0, t = at;
      for (; t !== null; )
        ao(t.alternate, t), t = t.return;
      at = null;
    }
  }
  function wa(t, l) {
    var e = t.timeoutHandle;
    e !== -1 && (t.timeoutHandle = -1, A0(e)), e = t.cancelPendingCommit, e !== null && (t.cancelPendingCommit = null, e()), Wc(), At = t, at = e = ue(t.current, null), ft = l, mt = 0, Al = null, Re = !1, Qa = Vl(t, l), Lc = !1, Za = Hl = Vc = fa = _e = Rt = 0, hl = Cu = null, wc = !1, (l & 8) !== 0 && (l |= l & 32);
    var a = t.entangledLanes;
    if (a !== 0)
      for (t = t.entanglements, a &= l; 0 < a; ) {
        var u = 31 - ct(a), n = 1 << u;
        l |= t[u], a &= ~n;
      }
    return ve = l, on(), e;
  }
  function zo(t, l) {
    lt = null, p.H = On, l === gu || l === xn ? (l = Zs(), mt = 3) : l === Gs ? (l = Zs(), mt = 4) : mt = l === Lr ? 8 : l !== null && typeof l == "object" && typeof l.then == "function" ? 6 : 1, Al = l, at === null && (Rt = 1, Hn(
      t,
      Ol(l, t.current)
    ));
  }
  function Mo() {
    var t = p.H;
    return p.H = On, t === null ? On : t;
  }
  function No() {
    var t = p.A;
    return p.A = u0, t;
  }
  function Fc() {
    Rt = 4, Re || (ft & 4194048) !== ft && Cl.current !== null || (Qa = !0), (_e & 134217727) === 0 && (fa & 134217727) === 0 || At === null || He(
      At,
      ft,
      Hl,
      !1
    );
  }
  function Pc(t, l, e) {
    var a = vt;
    vt |= 2;
    var u = Mo(), n = No();
    (At !== t || ft !== l) && (Zn = null, wa(t, l)), l = !1;
    var i = Rt;
    t: do
      try {
        if (mt !== 0 && at !== null) {
          var f = at, s = Al;
          switch (mt) {
            case 8:
              Wc(), i = 6;
              break t;
            case 3:
            case 2:
            case 9:
            case 6:
              Cl.current === null && (l = !0);
              var m = mt;
              if (mt = 0, Al = null, Ka(t, f, s, m), e && Qa) {
                i = 0;
                break t;
              }
              break;
            default:
              m = mt, mt = 0, Al = null, Ka(t, f, s, m);
          }
        }
        c0(), i = Rt;
        break;
      } catch (j) {
        zo(t, j);
      }
    while (!0);
    return l && t.shellSuspendCounter++, ce = ea = null, vt = a, p.H = u, p.A = n, at === null && (At = null, ft = 0, on()), i;
  }
  function c0() {
    for (; at !== null; ) Do(at);
  }
  function f0(t, l) {
    var e = vt;
    vt |= 2;
    var a = Mo(), u = No();
    At !== t || ft !== l ? (Zn = null, Qn = gl() + 500, wa(t, l)) : Qa = Vl(
      t,
      l
    );
    t: do
      try {
        if (mt !== 0 && at !== null) {
          l = at;
          var n = Al;
          l: switch (mt) {
            case 1:
              mt = 0, Al = null, Ka(t, l, n, 1);
              break;
            case 2:
            case 9:
              if (Xs(n)) {
                mt = 0, Al = null, Oo(l);
                break;
              }
              l = function() {
                mt !== 2 && mt !== 9 || At !== t || (mt = 7), Wl(t);
              }, n.then(l, l);
              break t;
            case 3:
              mt = 7;
              break t;
            case 4:
              mt = 5;
              break t;
            case 7:
              Xs(n) ? (mt = 0, Al = null, Oo(l)) : (mt = 0, Al = null, Ka(t, l, n, 7));
              break;
            case 5:
              var i = null;
              switch (at.tag) {
                case 26:
                  i = at.memoizedState;
                case 5:
                case 27:
                  var f = at;
                  if (!i || od(i)) {
                    mt = 0, Al = null;
                    var s = f.sibling;
                    if (s !== null) at = s;
                    else {
                      var m = f.return;
                      m !== null ? (at = m, Vn(m)) : at = null;
                    }
                    break l;
                  }
              }
              mt = 0, Al = null, Ka(t, l, n, 5);
              break;
            case 6:
              mt = 0, Al = null, Ka(t, l, n, 6);
              break;
            case 8:
              Wc(), Rt = 6;
              break t;
            default:
              throw Error(o(462));
          }
        }
        s0();
        break;
      } catch (j) {
        zo(t, j);
      }
    while (!0);
    return ce = ea = null, p.H = a, p.A = u, vt = e, at !== null ? 0 : (At = null, ft = 0, on(), Rt);
  }
  function s0() {
    for (; at !== null && !Fu(); )
      Do(at);
  }
  function Do(t) {
    var l = lo(t.alternate, t, ve);
    t.memoizedProps = t.pendingProps, l === null ? Vn(t) : at = l;
  }
  function Oo(t) {
    var l = t, e = l.alternate;
    switch (l.tag) {
      case 15:
      case 0:
        l = $r(
          e,
          l,
          l.pendingProps,
          l.type,
          void 0,
          ft
        );
        break;
      case 11:
        l = $r(
          e,
          l,
          l.pendingProps,
          l.type.render,
          l.ref,
          ft
        );
        break;
      case 5:
        hc(l);
      default:
        ao(e, l), l = at = Os(l, ve), l = lo(e, l, ve);
    }
    t.memoizedProps = t.pendingProps, l === null ? Vn(t) : at = l;
  }
  function Ka(t, l, e, a) {
    ce = ea = null, hc(l), Ba = null, zu = 0;
    var u = l.return;
    try {
      if (Ph(
        t,
        u,
        l,
        e,
        ft
      )) {
        Rt = 1, Hn(
          t,
          Ol(e, t.current)
        ), at = null;
        return;
      }
    } catch (n) {
      if (u !== null) throw at = u, n;
      Rt = 1, Hn(
        t,
        Ol(e, t.current)
      ), at = null;
      return;
    }
    l.flags & 32768 ? (yt || a === 1 ? t = !0 : Qa || (ft & 536870912) !== 0 ? t = !1 : (Re = t = !0, (a === 2 || a === 9 || a === 3 || a === 6) && (a = Cl.current, a !== null && a.tag === 13 && (a.flags |= 16384))), Ro(l, t)) : Vn(l);
  }
  function Vn(t) {
    var l = t;
    do {
      if ((l.flags & 32768) !== 0) {
        Ro(
          l,
          Re
        );
        return;
      }
      t = l.return;
      var e = t0(
        l.alternate,
        l,
        ve
      );
      if (e !== null) {
        at = e;
        return;
      }
      if (l = l.sibling, l !== null) {
        at = l;
        return;
      }
      at = l = t;
    } while (l !== null);
    Rt === 0 && (Rt = 5);
  }
  function Ro(t, l) {
    do {
      var e = l0(t.alternate, t);
      if (e !== null) {
        e.flags &= 32767, at = e;
        return;
      }
      if (e = t.return, e !== null && (e.flags |= 32768, e.subtreeFlags = 0, e.deletions = null), !l && (t = t.sibling, t !== null)) {
        at = t;
        return;
      }
      at = t = e;
    } while (t !== null);
    Rt = 6, at = null;
  }
  function _o(t, l, e, a, u, n, i, f, s) {
    t.cancelPendingCommit = null;
    do
      wn();
    while (It !== 0);
    if ((vt & 6) !== 0) throw Error(o(327));
    if (l !== null) {
      if (l === t.current) throw Error(o(177));
      if (n = l.lanes | l.childLanes, n |= Zi, dt(
        t,
        e,
        n,
        i,
        f,
        s
      ), t === At && (at = At = null, ft = 0), La = l, Ce = t, Va = e, Jc = n, kc = u, jo = a, (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? (t.callbackNode = null, t.callbackPriority = 0, h0(Ml, function() {
        return Bo(), null;
      })) : (t.callbackNode = null, t.callbackPriority = 0), a = (l.flags & 13878) !== 0, (l.subtreeFlags & 13878) !== 0 || a) {
        a = p.T, p.T = null, u = H.p, H.p = 2, i = vt, vt |= 4;
        try {
          e0(t, l, e);
        } finally {
          vt = i, H.p = u, p.T = a;
        }
      }
      It = 1, Uo(), Co(), Ho();
    }
  }
  function Uo() {
    if (It === 1) {
      It = 0;
      var t = Ce, l = La, e = (l.flags & 13878) !== 0;
      if ((l.subtreeFlags & 13878) !== 0 || e) {
        e = p.T, p.T = null;
        var a = H.p;
        H.p = 2;
        var u = vt;
        vt |= 4;
        try {
          vo(l, t);
          var n = of, i = xs(t.containerInfo), f = n.focusedElem, s = n.selectionRange;
          if (i !== f && f && f.ownerDocument && bs(
            f.ownerDocument.documentElement,
            f
          )) {
            if (s !== null && Bi(f)) {
              var m = s.start, j = s.end;
              if (j === void 0 && (j = m), "selectionStart" in f)
                f.selectionStart = m, f.selectionEnd = Math.min(
                  j,
                  f.value.length
                );
              else {
                var z = f.ownerDocument || document, g = z && z.defaultView || window;
                if (g.getSelection) {
                  var S = g.getSelection(), k = f.textContent.length, K = Math.min(s.start, k), bt = s.end === void 0 ? K : Math.min(s.end, k);
                  !S.extend && K > bt && (i = bt, bt = K, K = i);
                  var y = Ss(
                    f,
                    K
                  ), h = Ss(
                    f,
                    bt
                  );
                  if (y && h && (S.rangeCount !== 1 || S.anchorNode !== y.node || S.anchorOffset !== y.offset || S.focusNode !== h.node || S.focusOffset !== h.offset)) {
                    var v = z.createRange();
                    v.setStart(y.node, y.offset), S.removeAllRanges(), K > bt ? (S.addRange(v), S.extend(h.node, h.offset)) : (v.setEnd(h.node, h.offset), S.addRange(v));
                  }
                }
              }
            }
            for (z = [], S = f; S = S.parentNode; )
              S.nodeType === 1 && z.push({
                element: S,
                left: S.scrollLeft,
                top: S.scrollTop
              });
            for (typeof f.focus == "function" && f.focus(), f = 0; f < z.length; f++) {
              var T = z[f];
              T.element.scrollLeft = T.left, T.element.scrollTop = T.top;
            }
          }
          ai = !!rf, of = rf = null;
        } finally {
          vt = u, H.p = a, p.T = e;
        }
      }
      t.current = l, It = 2;
    }
  }
  function Co() {
    if (It === 2) {
      It = 0;
      var t = Ce, l = La, e = (l.flags & 8772) !== 0;
      if ((l.subtreeFlags & 8772) !== 0 || e) {
        e = p.T, p.T = null;
        var a = H.p;
        H.p = 2;
        var u = vt;
        vt |= 4;
        try {
          ro(t, l.alternate, l);
        } finally {
          vt = u, H.p = a, p.T = e;
        }
      }
      It = 3;
    }
  }
  function Ho() {
    if (It === 4 || It === 3) {
      It = 0, yi();
      var t = Ce, l = La, e = Va, a = jo;
      (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? It = 5 : (It = 0, La = Ce = null, qo(t, t.pendingLanes));
      var u = t.pendingLanes;
      if (u === 0 && (Ue = null), Bl(e), l = l.stateNode, Ft && typeof Ft.onCommitFiberRoot == "function")
        try {
          Ft.onCommitFiberRoot(
            Lt,
            l,
            void 0,
            (l.current.flags & 128) === 128
          );
        } catch {
        }
      if (a !== null) {
        l = p.T, u = H.p, H.p = 2, p.T = null;
        try {
          for (var n = t.onRecoverableError, i = 0; i < a.length; i++) {
            var f = a[i];
            n(f.value, {
              componentStack: f.stack
            });
          }
        } finally {
          p.T = l, H.p = u;
        }
      }
      (Va & 3) !== 0 && wn(), Wl(t), u = t.pendingLanes, (e & 4194090) !== 0 && (u & 42) !== 0 ? t === $c ? Hu++ : (Hu = 0, $c = t) : Hu = 0, qu(0);
    }
  }
  function qo(t, l) {
    (t.pooledCacheLanes &= l) === 0 && (l = t.pooledCache, l != null && (t.pooledCache = null, vu(l)));
  }
  function wn(t) {
    return Uo(), Co(), Ho(), Bo();
  }
  function Bo() {
    if (It !== 5) return !1;
    var t = Ce, l = Jc;
    Jc = 0;
    var e = Bl(Va), a = p.T, u = H.p;
    try {
      H.p = 32 > e ? 32 : e, p.T = null, e = kc, kc = null;
      var n = Ce, i = Va;
      if (It = 0, La = Ce = null, Va = 0, (vt & 6) !== 0) throw Error(o(331));
      var f = vt;
      if (vt |= 4, xo(n.current), go(
        n,
        n.current,
        i,
        e
      ), vt = f, qu(0, !1), Ft && typeof Ft.onPostCommitFiberRoot == "function")
        try {
          Ft.onPostCommitFiberRoot(Lt, n);
        } catch {
        }
      return !0;
    } finally {
      H.p = u, p.T = a, qo(t, l);
    }
  }
  function Yo(t, l, e) {
    l = Ol(e, l), l = Mc(t.stateNode, l, 2), t = Ae(t, l, 2), t !== null && (tt(t, 2), Wl(t));
  }
  function pt(t, l, e) {
    if (t.tag === 3)
      Yo(t, t, e);
    else
      for (; l !== null; ) {
        if (l.tag === 3) {
          Yo(
            l,
            t,
            e
          );
          break;
        } else if (l.tag === 1) {
          var a = l.stateNode;
          if (typeof l.type.getDerivedStateFromError == "function" || typeof a.componentDidCatch == "function" && (Ue === null || !Ue.has(a))) {
            t = Ol(e, t), e = Qr(2), a = Ae(l, e, 2), a !== null && (Zr(
              e,
              a,
              l,
              t
            ), tt(a, 2), Wl(a));
            break;
          }
        }
        l = l.return;
      }
  }
  function Ic(t, l, e) {
    var a = t.pingCache;
    if (a === null) {
      a = t.pingCache = new n0();
      var u = /* @__PURE__ */ new Set();
      a.set(l, u);
    } else
      u = a.get(l), u === void 0 && (u = /* @__PURE__ */ new Set(), a.set(l, u));
    u.has(e) || (Lc = !0, u.add(e), t = r0.bind(null, t, l, e), l.then(t, t));
  }
  function r0(t, l, e) {
    var a = t.pingCache;
    a !== null && a.delete(l), t.pingedLanes |= t.suspendedLanes & e, t.warmLanes &= ~e, At === t && (ft & e) === e && (Rt === 4 || Rt === 3 && (ft & 62914560) === ft && 300 > gl() - Kc ? (vt & 2) === 0 && wa(t, 0) : Vc |= e, Za === ft && (Za = 0)), Wl(t);
  }
  function Go(t, l) {
    l === 0 && (l = X()), t = Ma(t, l), t !== null && (tt(t, l), Wl(t));
  }
  function o0(t) {
    var l = t.memoizedState, e = 0;
    l !== null && (e = l.retryLane), Go(t, e);
  }
  function d0(t, l) {
    var e = 0;
    switch (t.tag) {
      case 13:
        var a = t.stateNode, u = t.memoizedState;
        u !== null && (e = u.retryLane);
        break;
      case 19:
        a = t.stateNode;
        break;
      case 22:
        a = t.stateNode._retryCache;
        break;
      default:
        throw Error(o(314));
    }
    a !== null && a.delete(l), Go(t, e);
  }
  function h0(t, l) {
    return oa(t, l);
  }
  var Kn = null, Ja = null, tf = !1, Jn = !1, lf = !1, sa = 0;
  function Wl(t) {
    t !== Ja && t.next === null && (Ja === null ? Kn = Ja = t : Ja = Ja.next = t), Jn = !0, tf || (tf = !0, v0());
  }
  function qu(t, l) {
    if (!lf && Jn) {
      lf = !0;
      do
        for (var e = !1, a = Kn; a !== null; ) {
          if (t !== 0) {
            var u = a.pendingLanes;
            if (u === 0) var n = 0;
            else {
              var i = a.suspendedLanes, f = a.pingedLanes;
              n = (1 << 31 - ct(42 | t) + 1) - 1, n &= u & ~(i & ~f), n = n & 201326741 ? n & 201326741 | 1 : n ? n | 2 : 0;
            }
            n !== 0 && (e = !0, Lo(a, n));
          } else
            n = ft, n = da(
              a,
              a === At ? n : 0,
              a.cancelPendingCommit !== null || a.timeoutHandle !== -1
            ), (n & 3) === 0 || Vl(a, n) || (e = !0, Lo(a, n));
          a = a.next;
        }
      while (e);
      lf = !1;
    }
  }
  function y0() {
    Xo();
  }
  function Xo() {
    Jn = tf = !1;
    var t = 0;
    sa !== 0 && (T0() && (t = sa), sa = 0);
    for (var l = gl(), e = null, a = Kn; a !== null; ) {
      var u = a.next, n = Qo(a, l);
      n === 0 ? (a.next = null, e === null ? Kn = u : e.next = u, u === null && (Ja = e)) : (e = a, (t !== 0 || (n & 3) !== 0) && (Jn = !0)), a = u;
    }
    qu(t);
  }
  function Qo(t, l) {
    for (var e = t.suspendedLanes, a = t.pingedLanes, u = t.expirationTimes, n = t.pendingLanes & -62914561; 0 < n; ) {
      var i = 31 - ct(n), f = 1 << i, s = u[i];
      s === -1 ? ((f & e) === 0 || (f & a) !== 0) && (u[i] = we(f, l)) : s <= l && (t.expiredLanes |= f), n &= ~f;
    }
    if (l = At, e = ft, e = da(
      t,
      t === l ? e : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a = t.callbackNode, e === 0 || t === l && (mt === 2 || mt === 9) || t.cancelPendingCommit !== null)
      return a !== null && a !== null && Pl(a), t.callbackNode = null, t.callbackPriority = 0;
    if ((e & 3) === 0 || Vl(t, e)) {
      if (l = e & -e, l === t.callbackPriority) return l;
      switch (a !== null && Pl(a), Bl(e)) {
        case 2:
        case 8:
          e = Il;
          break;
        case 32:
          e = Ml;
          break;
        case 268435456:
          e = be;
          break;
        default:
          e = Ml;
      }
      return a = Zo.bind(null, t), e = oa(e, a), t.callbackPriority = l, t.callbackNode = e, l;
    }
    return a !== null && a !== null && Pl(a), t.callbackPriority = 2, t.callbackNode = null, 2;
  }
  function Zo(t, l) {
    if (It !== 0 && It !== 5)
      return t.callbackNode = null, t.callbackPriority = 0, null;
    var e = t.callbackNode;
    if (wn() && t.callbackNode !== e)
      return null;
    var a = ft;
    return a = da(
      t,
      t === At ? a : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a === 0 ? null : (Ao(t, a, l), Qo(t, gl()), t.callbackNode != null && t.callbackNode === e ? Zo.bind(null, t) : null);
  }
  function Lo(t, l) {
    if (wn()) return null;
    Ao(t, l, !0);
  }
  function v0() {
    E0(function() {
      (vt & 6) !== 0 ? oa(
        Ze,
        y0
      ) : Xo();
    });
  }
  function ef() {
    return sa === 0 && (sa = E()), sa;
  }
  function Vo(t) {
    return t == null || typeof t == "symbol" || typeof t == "boolean" ? null : typeof t == "function" ? t : an("" + t);
  }
  function wo(t, l) {
    var e = l.ownerDocument.createElement("input");
    return e.name = l.name, e.value = l.value, t.id && e.setAttribute("form", t.id), l.parentNode.insertBefore(e, l), t = new FormData(t), e.parentNode.removeChild(e), t;
  }
  function m0(t, l, e, a, u) {
    if (l === "submit" && e && e.stateNode === u) {
      var n = Vo(
        (u[sl] || null).action
      ), i = a.submitter;
      i && (l = (l = i[sl] || null) ? Vo(l.formAction) : i.getAttribute("formAction"), l !== null && (n = l, i = null));
      var f = new fn(
        "action",
        "action",
        null,
        a,
        u
      );
      t.push({
        event: f,
        listeners: [
          {
            instance: null,
            listener: function() {
              if (a.defaultPrevented) {
                if (sa !== 0) {
                  var s = i ? wo(u, i) : new FormData(u);
                  jc(
                    e,
                    {
                      pending: !0,
                      data: s,
                      method: u.method,
                      action: n
                    },
                    null,
                    s
                  );
                }
              } else
                typeof n == "function" && (f.preventDefault(), s = i ? wo(u, i) : new FormData(u), jc(
                  e,
                  {
                    pending: !0,
                    data: s,
                    method: u.method,
                    action: n
                  },
                  n,
                  s
                ));
            },
            currentTarget: u
          }
        ]
      });
    }
  }
  for (var af = 0; af < Qi.length; af++) {
    var uf = Qi[af], g0 = uf.toLowerCase(), S0 = uf[0].toUpperCase() + uf.slice(1);
    Yl(
      g0,
      "on" + S0
    );
  }
  Yl(Ts, "onAnimationEnd"), Yl(As, "onAnimationIteration"), Yl(Es, "onAnimationStart"), Yl("dblclick", "onDoubleClick"), Yl("focusin", "onFocus"), Yl("focusout", "onBlur"), Yl(Hh, "onTransitionRun"), Yl(qh, "onTransitionStart"), Yl(Bh, "onTransitionCancel"), Yl(zs, "onTransitionEnd"), ga("onMouseEnter", ["mouseout", "mouseover"]), ga("onMouseLeave", ["mouseout", "mouseover"]), ga("onPointerEnter", ["pointerout", "pointerover"]), ga("onPointerLeave", ["pointerout", "pointerover"]), Je(
    "onChange",
    "change click focusin focusout input keydown keyup selectionchange".split(" ")
  ), Je(
    "onSelect",
    "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(
      " "
    )
  ), Je("onBeforeInput", [
    "compositionend",
    "keypress",
    "textInput",
    "paste"
  ]), Je(
    "onCompositionEnd",
    "compositionend focusout keydown keypress keyup mousedown".split(" ")
  ), Je(
    "onCompositionStart",
    "compositionstart focusout keydown keypress keyup mousedown".split(" ")
  ), Je(
    "onCompositionUpdate",
    "compositionupdate focusout keydown keypress keyup mousedown".split(" ")
  );
  var Bu = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(
    " "
  ), b0 = new Set(
    "beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(Bu)
  );
  function Ko(t, l) {
    l = (l & 4) !== 0;
    for (var e = 0; e < t.length; e++) {
      var a = t[e], u = a.event;
      a = a.listeners;
      t: {
        var n = void 0;
        if (l)
          for (var i = a.length - 1; 0 <= i; i--) {
            var f = a[i], s = f.instance, m = f.currentTarget;
            if (f = f.listener, s !== n && u.isPropagationStopped())
              break t;
            n = f, u.currentTarget = m;
            try {
              n(u);
            } catch (j) {
              Cn(j);
            }
            u.currentTarget = null, n = s;
          }
        else
          for (i = 0; i < a.length; i++) {
            if (f = a[i], s = f.instance, m = f.currentTarget, f = f.listener, s !== n && u.isPropagationStopped())
              break t;
            n = f, u.currentTarget = m;
            try {
              n(u);
            } catch (j) {
              Cn(j);
            }
            u.currentTarget = null, n = s;
          }
      }
    }
  }
  function ut(t, l) {
    var e = l[gi];
    e === void 0 && (e = l[gi] = /* @__PURE__ */ new Set());
    var a = t + "__bubble";
    e.has(a) || (Jo(l, t, 2, !1), e.add(a));
  }
  function nf(t, l, e) {
    var a = 0;
    l && (a |= 4), Jo(
      e,
      t,
      a,
      l
    );
  }
  var kn = "_reactListening" + Math.random().toString(36).slice(2);
  function cf(t) {
    if (!t[kn]) {
      t[kn] = !0, Gf.forEach(function(e) {
        e !== "selectionchange" && (b0.has(e) || nf(e, !1, t), nf(e, !0, t));
      });
      var l = t.nodeType === 9 ? t : t.ownerDocument;
      l === null || l[kn] || (l[kn] = !0, nf("selectionchange", !1, l));
    }
  }
  function Jo(t, l, e, a) {
    switch (gd(l)) {
      case 2:
        var u = K0;
        break;
      case 8:
        u = J0;
        break;
      default:
        u = pf;
    }
    e = u.bind(
      null,
      l,
      e,
      t
    ), u = void 0, !Ni || l !== "touchstart" && l !== "touchmove" && l !== "wheel" || (u = !0), a ? u !== void 0 ? t.addEventListener(l, e, {
      capture: !0,
      passive: u
    }) : t.addEventListener(l, e, !0) : u !== void 0 ? t.addEventListener(l, e, {
      passive: u
    }) : t.addEventListener(l, e, !1);
  }
  function ff(t, l, e, a, u) {
    var n = a;
    if ((l & 1) === 0 && (l & 2) === 0 && a !== null)
      t: for (; ; ) {
        if (a === null) return;
        var i = a.tag;
        if (i === 3 || i === 4) {
          var f = a.stateNode.containerInfo;
          if (f === u) break;
          if (i === 4)
            for (i = a.return; i !== null; ) {
              var s = i.tag;
              if ((s === 3 || s === 4) && i.stateNode.containerInfo === u)
                return;
              i = i.return;
            }
          for (; f !== null; ) {
            if (i = ya(f), i === null) return;
            if (s = i.tag, s === 5 || s === 6 || s === 26 || s === 27) {
              a = n = i;
              continue t;
            }
            f = f.parentNode;
          }
        }
        a = a.return;
      }
    If(function() {
      var m = n, j = zi(e), z = [];
      t: {
        var g = Ms.get(t);
        if (g !== void 0) {
          var S = fn, k = t;
          switch (t) {
            case "keypress":
              if (nn(e) === 0) break t;
            case "keydown":
            case "keyup":
              S = hh;
              break;
            case "focusin":
              k = "focus", S = _i;
              break;
            case "focusout":
              k = "blur", S = _i;
              break;
            case "beforeblur":
            case "afterblur":
              S = _i;
              break;
            case "click":
              if (e.button === 2) break t;
            case "auxclick":
            case "dblclick":
            case "mousedown":
            case "mousemove":
            case "mouseup":
            case "mouseout":
            case "mouseover":
            case "contextmenu":
              S = es;
              break;
            case "drag":
            case "dragend":
            case "dragenter":
            case "dragexit":
            case "dragleave":
            case "dragover":
            case "dragstart":
            case "drop":
              S = lh;
              break;
            case "touchcancel":
            case "touchend":
            case "touchmove":
            case "touchstart":
              S = mh;
              break;
            case Ts:
            case As:
            case Es:
              S = uh;
              break;
            case zs:
              S = Sh;
              break;
            case "scroll":
            case "scrollend":
              S = Id;
              break;
            case "wheel":
              S = xh;
              break;
            case "copy":
            case "cut":
            case "paste":
              S = ih;
              break;
            case "gotpointercapture":
            case "lostpointercapture":
            case "pointercancel":
            case "pointerdown":
            case "pointermove":
            case "pointerout":
            case "pointerover":
            case "pointerup":
              S = us;
              break;
            case "toggle":
            case "beforetoggle":
              S = jh;
          }
          var K = (l & 4) !== 0, bt = !K && (t === "scroll" || t === "scrollend"), y = K ? g !== null ? g + "Capture" : null : g;
          K = [];
          for (var h = m, v; h !== null; ) {
            var T = h;
            if (v = T.stateNode, T = T.tag, T !== 5 && T !== 26 && T !== 27 || v === null || y === null || (T = eu(h, y), T != null && K.push(
              Yu(h, T, v)
            )), bt) break;
            h = h.return;
          }
          0 < K.length && (g = new S(
            g,
            k,
            null,
            e,
            j
          ), z.push({ event: g, listeners: K }));
        }
      }
      if ((l & 7) === 0) {
        t: {
          if (g = t === "mouseover" || t === "pointerover", S = t === "mouseout" || t === "pointerout", g && e !== Ei && (k = e.relatedTarget || e.fromElement) && (ya(k) || k[ha]))
            break t;
          if ((S || g) && (g = j.window === j ? j : (g = j.ownerDocument) ? g.defaultView || g.parentWindow : window, S ? (k = e.relatedTarget || e.toElement, S = m, k = k ? ya(k) : null, k !== null && (bt = D(k), K = k.tag, k !== bt || K !== 5 && K !== 27 && K !== 6) && (k = null)) : (S = null, k = m), S !== k)) {
            if (K = es, T = "onMouseLeave", y = "onMouseEnter", h = "mouse", (t === "pointerout" || t === "pointerover") && (K = us, T = "onPointerLeave", y = "onPointerEnter", h = "pointer"), bt = S == null ? g : lu(S), v = k == null ? g : lu(k), g = new K(
              T,
              h + "leave",
              S,
              e,
              j
            ), g.target = bt, g.relatedTarget = v, T = null, ya(j) === m && (K = new K(
              y,
              h + "enter",
              k,
              e,
              j
            ), K.target = v, K.relatedTarget = bt, T = K), bt = T, S && k)
              l: {
                for (K = S, y = k, h = 0, v = K; v; v = ka(v))
                  h++;
                for (v = 0, T = y; T; T = ka(T))
                  v++;
                for (; 0 < h - v; )
                  K = ka(K), h--;
                for (; 0 < v - h; )
                  y = ka(y), v--;
                for (; h--; ) {
                  if (K === y || y !== null && K === y.alternate)
                    break l;
                  K = ka(K), y = ka(y);
                }
                K = null;
              }
            else K = null;
            S !== null && ko(
              z,
              g,
              S,
              K,
              !1
            ), k !== null && bt !== null && ko(
              z,
              bt,
              k,
              K,
              !0
            );
          }
        }
        t: {
          if (g = m ? lu(m) : window, S = g.nodeName && g.nodeName.toLowerCase(), S === "select" || S === "input" && g.type === "file")
            var Q = ds;
          else if (rs(g))
            if (hs)
              Q = _h;
            else {
              Q = Oh;
              var et = Dh;
            }
          else
            S = g.nodeName, !S || S.toLowerCase() !== "input" || g.type !== "checkbox" && g.type !== "radio" ? m && Ai(m.elementType) && (Q = ds) : Q = Rh;
          if (Q && (Q = Q(t, m))) {
            os(
              z,
              Q,
              e,
              j
            );
            break t;
          }
          et && et(t, g, m), t === "focusout" && m && g.type === "number" && m.memoizedProps.value != null && Ti(g, "number", g.value);
        }
        switch (et = m ? lu(m) : window, t) {
          case "focusin":
            (rs(et) || et.contentEditable === "true") && (Aa = et, Yi = m, ru = null);
            break;
          case "focusout":
            ru = Yi = Aa = null;
            break;
          case "mousedown":
            Gi = !0;
            break;
          case "contextmenu":
          case "mouseup":
          case "dragend":
            Gi = !1, ps(z, e, j);
            break;
          case "selectionchange":
            if (Ch) break;
          case "keydown":
          case "keyup":
            ps(z, e, j);
        }
        var Z;
        if (Ci)
          t: {
            switch (t) {
              case "compositionstart":
                var J = "onCompositionStart";
                break t;
              case "compositionend":
                J = "onCompositionEnd";
                break t;
              case "compositionupdate":
                J = "onCompositionUpdate";
                break t;
            }
            J = void 0;
          }
        else
          Ta ? fs(t, e) && (J = "onCompositionEnd") : t === "keydown" && e.keyCode === 229 && (J = "onCompositionStart");
        J && (ns && e.locale !== "ko" && (Ta || J !== "onCompositionStart" ? J === "onCompositionEnd" && Ta && (Z = ts()) : (xe = j, Di = "value" in xe ? xe.value : xe.textContent, Ta = !0)), et = $n(m, J), 0 < et.length && (J = new as(
          J,
          t,
          null,
          e,
          j
        ), z.push({ event: J, listeners: et }), Z ? J.data = Z : (Z = ss(e), Z !== null && (J.data = Z)))), (Z = Ah ? Eh(t, e) : zh(t, e)) && (J = $n(m, "onBeforeInput"), 0 < J.length && (et = new as(
          "onBeforeInput",
          "beforeinput",
          null,
          e,
          j
        ), z.push({
          event: et,
          listeners: J
        }), et.data = Z)), m0(
          z,
          t,
          m,
          e,
          j
        );
      }
      Ko(z, l);
    });
  }
  function Yu(t, l, e) {
    return {
      instance: t,
      listener: l,
      currentTarget: e
    };
  }
  function $n(t, l) {
    for (var e = l + "Capture", a = []; t !== null; ) {
      var u = t, n = u.stateNode;
      if (u = u.tag, u !== 5 && u !== 26 && u !== 27 || n === null || (u = eu(t, e), u != null && a.unshift(
        Yu(t, u, n)
      ), u = eu(t, l), u != null && a.push(
        Yu(t, u, n)
      )), t.tag === 3) return a;
      t = t.return;
    }
    return [];
  }
  function ka(t) {
    if (t === null) return null;
    do
      t = t.return;
    while (t && t.tag !== 5 && t.tag !== 27);
    return t || null;
  }
  function ko(t, l, e, a, u) {
    for (var n = l._reactName, i = []; e !== null && e !== a; ) {
      var f = e, s = f.alternate, m = f.stateNode;
      if (f = f.tag, s !== null && s === a) break;
      f !== 5 && f !== 26 && f !== 27 || m === null || (s = m, u ? (m = eu(e, n), m != null && i.unshift(
        Yu(e, m, s)
      )) : u || (m = eu(e, n), m != null && i.push(
        Yu(e, m, s)
      ))), e = e.return;
    }
    i.length !== 0 && t.push({ event: l, listeners: i });
  }
  var x0 = /\r\n?/g, p0 = /\u0000|\uFFFD/g;
  function $o(t) {
    return (typeof t == "string" ? t : "" + t).replace(x0, `
`).replace(p0, "");
  }
  function Wo(t, l) {
    return l = $o(l), $o(t) === l;
  }
  function Wn() {
  }
  function St(t, l, e, a, u, n) {
    switch (e) {
      case "children":
        typeof a == "string" ? l === "body" || l === "textarea" && a === "" || xa(t, a) : (typeof a == "number" || typeof a == "bigint") && l !== "body" && xa(t, "" + a);
        break;
      case "className":
        tn(t, "class", a);
        break;
      case "tabIndex":
        tn(t, "tabindex", a);
        break;
      case "dir":
      case "role":
      case "viewBox":
      case "width":
      case "height":
        tn(t, e, a);
        break;
      case "style":
        Ff(t, a, n);
        break;
      case "data":
        if (l !== "object") {
          tn(t, "data", a);
          break;
        }
      case "src":
      case "href":
        if (a === "" && (l !== "a" || e !== "href")) {
          t.removeAttribute(e);
          break;
        }
        if (a == null || typeof a == "function" || typeof a == "symbol" || typeof a == "boolean") {
          t.removeAttribute(e);
          break;
        }
        a = an("" + a), t.setAttribute(e, a);
        break;
      case "action":
      case "formAction":
        if (typeof a == "function") {
          t.setAttribute(
            e,
            "javascript:throw new Error('A React form was unexpectedly submitted. If you called form.submit() manually, consider using form.requestSubmit() instead. If you\\'re trying to use event.stopPropagation() in a submit event handler, consider also calling event.preventDefault().')"
          );
          break;
        } else
          typeof n == "function" && (e === "formAction" ? (l !== "input" && St(t, l, "name", u.name, u, null), St(
            t,
            l,
            "formEncType",
            u.formEncType,
            u,
            null
          ), St(
            t,
            l,
            "formMethod",
            u.formMethod,
            u,
            null
          ), St(
            t,
            l,
            "formTarget",
            u.formTarget,
            u,
            null
          )) : (St(t, l, "encType", u.encType, u, null), St(t, l, "method", u.method, u, null), St(t, l, "target", u.target, u, null)));
        if (a == null || typeof a == "symbol" || typeof a == "boolean") {
          t.removeAttribute(e);
          break;
        }
        a = an("" + a), t.setAttribute(e, a);
        break;
      case "onClick":
        a != null && (t.onclick = Wn);
        break;
      case "onScroll":
        a != null && ut("scroll", t);
        break;
      case "onScrollEnd":
        a != null && ut("scrollend", t);
        break;
      case "dangerouslySetInnerHTML":
        if (a != null) {
          if (typeof a != "object" || !("__html" in a))
            throw Error(o(61));
          if (e = a.__html, e != null) {
            if (u.children != null) throw Error(o(60));
            t.innerHTML = e;
          }
        }
        break;
      case "multiple":
        t.multiple = a && typeof a != "function" && typeof a != "symbol";
        break;
      case "muted":
        t.muted = a && typeof a != "function" && typeof a != "symbol";
        break;
      case "suppressContentEditableWarning":
      case "suppressHydrationWarning":
      case "defaultValue":
      case "defaultChecked":
      case "innerHTML":
      case "ref":
        break;
      case "autoFocus":
        break;
      case "xlinkHref":
        if (a == null || typeof a == "function" || typeof a == "boolean" || typeof a == "symbol") {
          t.removeAttribute("xlink:href");
          break;
        }
        e = an("" + a), t.setAttributeNS(
          "http://www.w3.org/1999/xlink",
          "xlink:href",
          e
        );
        break;
      case "contentEditable":
      case "spellCheck":
      case "draggable":
      case "value":
      case "autoReverse":
      case "externalResourcesRequired":
      case "focusable":
      case "preserveAlpha":
        a != null && typeof a != "function" && typeof a != "symbol" ? t.setAttribute(e, "" + a) : t.removeAttribute(e);
        break;
      case "inert":
      case "allowFullScreen":
      case "async":
      case "autoPlay":
      case "controls":
      case "default":
      case "defer":
      case "disabled":
      case "disablePictureInPicture":
      case "disableRemotePlayback":
      case "formNoValidate":
      case "hidden":
      case "loop":
      case "noModule":
      case "noValidate":
      case "open":
      case "playsInline":
      case "readOnly":
      case "required":
      case "reversed":
      case "scoped":
      case "seamless":
      case "itemScope":
        a && typeof a != "function" && typeof a != "symbol" ? t.setAttribute(e, "") : t.removeAttribute(e);
        break;
      case "capture":
      case "download":
        a === !0 ? t.setAttribute(e, "") : a !== !1 && a != null && typeof a != "function" && typeof a != "symbol" ? t.setAttribute(e, a) : t.removeAttribute(e);
        break;
      case "cols":
      case "rows":
      case "size":
      case "span":
        a != null && typeof a != "function" && typeof a != "symbol" && !isNaN(a) && 1 <= a ? t.setAttribute(e, a) : t.removeAttribute(e);
        break;
      case "rowSpan":
      case "start":
        a == null || typeof a == "function" || typeof a == "symbol" || isNaN(a) ? t.removeAttribute(e) : t.setAttribute(e, a);
        break;
      case "popover":
        ut("beforetoggle", t), ut("toggle", t), Iu(t, "popover", a);
        break;
      case "xlinkActuate":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:actuate",
          a
        );
        break;
      case "xlinkArcrole":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:arcrole",
          a
        );
        break;
      case "xlinkRole":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:role",
          a
        );
        break;
      case "xlinkShow":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:show",
          a
        );
        break;
      case "xlinkTitle":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:title",
          a
        );
        break;
      case "xlinkType":
        ee(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:type",
          a
        );
        break;
      case "xmlBase":
        ee(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:base",
          a
        );
        break;
      case "xmlLang":
        ee(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:lang",
          a
        );
        break;
      case "xmlSpace":
        ee(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:space",
          a
        );
        break;
      case "is":
        Iu(t, "is", a);
        break;
      case "innerText":
      case "textContent":
        break;
      default:
        (!(2 < e.length) || e[0] !== "o" && e[0] !== "O" || e[1] !== "n" && e[1] !== "N") && (e = Fd.get(e) || e, Iu(t, e, a));
    }
  }
  function sf(t, l, e, a, u, n) {
    switch (e) {
      case "style":
        Ff(t, a, n);
        break;
      case "dangerouslySetInnerHTML":
        if (a != null) {
          if (typeof a != "object" || !("__html" in a))
            throw Error(o(61));
          if (e = a.__html, e != null) {
            if (u.children != null) throw Error(o(60));
            t.innerHTML = e;
          }
        }
        break;
      case "children":
        typeof a == "string" ? xa(t, a) : (typeof a == "number" || typeof a == "bigint") && xa(t, "" + a);
        break;
      case "onScroll":
        a != null && ut("scroll", t);
        break;
      case "onScrollEnd":
        a != null && ut("scrollend", t);
        break;
      case "onClick":
        a != null && (t.onclick = Wn);
        break;
      case "suppressContentEditableWarning":
      case "suppressHydrationWarning":
      case "innerHTML":
      case "ref":
        break;
      case "innerText":
      case "textContent":
        break;
      default:
        if (!Xf.hasOwnProperty(e))
          t: {
            if (e[0] === "o" && e[1] === "n" && (u = e.endsWith("Capture"), l = e.slice(2, u ? e.length - 7 : void 0), n = t[sl] || null, n = n != null ? n[e] : null, typeof n == "function" && t.removeEventListener(l, n, u), typeof a == "function")) {
              typeof n != "function" && n !== null && (e in t ? t[e] = null : t.hasAttribute(e) && t.removeAttribute(e)), t.addEventListener(l, a, u);
              break t;
            }
            e in t ? t[e] = a : a === !0 ? t.setAttribute(e, "") : Iu(t, e, a);
          }
    }
  }
  function tl(t, l, e) {
    switch (l) {
      case "div":
      case "span":
      case "svg":
      case "path":
      case "a":
      case "g":
      case "p":
      case "li":
        break;
      case "img":
        ut("error", t), ut("load", t);
        var a = !1, u = !1, n;
        for (n in e)
          if (e.hasOwnProperty(n)) {
            var i = e[n];
            if (i != null)
              switch (n) {
                case "src":
                  a = !0;
                  break;
                case "srcSet":
                  u = !0;
                  break;
                case "children":
                case "dangerouslySetInnerHTML":
                  throw Error(o(137, l));
                default:
                  St(t, l, n, i, e, null);
              }
          }
        u && St(t, l, "srcSet", e.srcSet, e, null), a && St(t, l, "src", e.src, e, null);
        return;
      case "input":
        ut("invalid", t);
        var f = n = i = u = null, s = null, m = null;
        for (a in e)
          if (e.hasOwnProperty(a)) {
            var j = e[a];
            if (j != null)
              switch (a) {
                case "name":
                  u = j;
                  break;
                case "type":
                  i = j;
                  break;
                case "checked":
                  s = j;
                  break;
                case "defaultChecked":
                  m = j;
                  break;
                case "value":
                  n = j;
                  break;
                case "defaultValue":
                  f = j;
                  break;
                case "children":
                case "dangerouslySetInnerHTML":
                  if (j != null)
                    throw Error(o(137, l));
                  break;
                default:
                  St(t, l, a, j, e, null);
              }
          }
        Jf(
          t,
          n,
          f,
          s,
          m,
          i,
          u,
          !1
        ), ln(t);
        return;
      case "select":
        ut("invalid", t), a = i = n = null;
        for (u in e)
          if (e.hasOwnProperty(u) && (f = e[u], f != null))
            switch (u) {
              case "value":
                n = f;
                break;
              case "defaultValue":
                i = f;
                break;
              case "multiple":
                a = f;
              default:
                St(t, l, u, f, e, null);
            }
        l = n, e = i, t.multiple = !!a, l != null ? ba(t, !!a, l, !1) : e != null && ba(t, !!a, e, !0);
        return;
      case "textarea":
        ut("invalid", t), n = u = a = null;
        for (i in e)
          if (e.hasOwnProperty(i) && (f = e[i], f != null))
            switch (i) {
              case "value":
                a = f;
                break;
              case "defaultValue":
                u = f;
                break;
              case "children":
                n = f;
                break;
              case "dangerouslySetInnerHTML":
                if (f != null) throw Error(o(91));
                break;
              default:
                St(t, l, i, f, e, null);
            }
        $f(t, a, u, n), ln(t);
        return;
      case "option":
        for (s in e)
          e.hasOwnProperty(s) && (a = e[s], a != null) && (s === "selected" ? t.selected = a && typeof a != "function" && typeof a != "symbol" : St(t, l, s, a, e, null));
        return;
      case "dialog":
        ut("beforetoggle", t), ut("toggle", t), ut("cancel", t), ut("close", t);
        break;
      case "iframe":
      case "object":
        ut("load", t);
        break;
      case "video":
      case "audio":
        for (a = 0; a < Bu.length; a++)
          ut(Bu[a], t);
        break;
      case "image":
        ut("error", t), ut("load", t);
        break;
      case "details":
        ut("toggle", t);
        break;
      case "embed":
      case "source":
      case "link":
        ut("error", t), ut("load", t);
      case "area":
      case "base":
      case "br":
      case "col":
      case "hr":
      case "keygen":
      case "meta":
      case "param":
      case "track":
      case "wbr":
      case "menuitem":
        for (m in e)
          if (e.hasOwnProperty(m) && (a = e[m], a != null))
            switch (m) {
              case "children":
              case "dangerouslySetInnerHTML":
                throw Error(o(137, l));
              default:
                St(t, l, m, a, e, null);
            }
        return;
      default:
        if (Ai(l)) {
          for (j in e)
            e.hasOwnProperty(j) && (a = e[j], a !== void 0 && sf(
              t,
              l,
              j,
              a,
              e,
              void 0
            ));
          return;
        }
    }
    for (f in e)
      e.hasOwnProperty(f) && (a = e[f], a != null && St(t, l, f, a, e, null));
  }
  function j0(t, l, e, a) {
    switch (l) {
      case "div":
      case "span":
      case "svg":
      case "path":
      case "a":
      case "g":
      case "p":
      case "li":
        break;
      case "input":
        var u = null, n = null, i = null, f = null, s = null, m = null, j = null;
        for (S in e) {
          var z = e[S];
          if (e.hasOwnProperty(S) && z != null)
            switch (S) {
              case "checked":
                break;
              case "value":
                break;
              case "defaultValue":
                s = z;
              default:
                a.hasOwnProperty(S) || St(t, l, S, null, a, z);
            }
        }
        for (var g in a) {
          var S = a[g];
          if (z = e[g], a.hasOwnProperty(g) && (S != null || z != null))
            switch (g) {
              case "type":
                n = S;
                break;
              case "name":
                u = S;
                break;
              case "checked":
                m = S;
                break;
              case "defaultChecked":
                j = S;
                break;
              case "value":
                i = S;
                break;
              case "defaultValue":
                f = S;
                break;
              case "children":
              case "dangerouslySetInnerHTML":
                if (S != null)
                  throw Error(o(137, l));
                break;
              default:
                S !== z && St(
                  t,
                  l,
                  g,
                  S,
                  a,
                  z
                );
            }
        }
        ji(
          t,
          i,
          f,
          s,
          m,
          j,
          n,
          u
        );
        return;
      case "select":
        S = i = f = g = null;
        for (n in e)
          if (s = e[n], e.hasOwnProperty(n) && s != null)
            switch (n) {
              case "value":
                break;
              case "multiple":
                S = s;
              default:
                a.hasOwnProperty(n) || St(
                  t,
                  l,
                  n,
                  null,
                  a,
                  s
                );
            }
        for (u in a)
          if (n = a[u], s = e[u], a.hasOwnProperty(u) && (n != null || s != null))
            switch (u) {
              case "value":
                g = n;
                break;
              case "defaultValue":
                f = n;
                break;
              case "multiple":
                i = n;
              default:
                n !== s && St(
                  t,
                  l,
                  u,
                  n,
                  a,
                  s
                );
            }
        l = f, e = i, a = S, g != null ? ba(t, !!e, g, !1) : !!a != !!e && (l != null ? ba(t, !!e, l, !0) : ba(t, !!e, e ? [] : "", !1));
        return;
      case "textarea":
        S = g = null;
        for (f in e)
          if (u = e[f], e.hasOwnProperty(f) && u != null && !a.hasOwnProperty(f))
            switch (f) {
              case "value":
                break;
              case "children":
                break;
              default:
                St(t, l, f, null, a, u);
            }
        for (i in a)
          if (u = a[i], n = e[i], a.hasOwnProperty(i) && (u != null || n != null))
            switch (i) {
              case "value":
                g = u;
                break;
              case "defaultValue":
                S = u;
                break;
              case "children":
                break;
              case "dangerouslySetInnerHTML":
                if (u != null) throw Error(o(91));
                break;
              default:
                u !== n && St(t, l, i, u, a, n);
            }
        kf(t, g, S);
        return;
      case "option":
        for (var k in e)
          g = e[k], e.hasOwnProperty(k) && g != null && !a.hasOwnProperty(k) && (k === "selected" ? t.selected = !1 : St(
            t,
            l,
            k,
            null,
            a,
            g
          ));
        for (s in a)
          g = a[s], S = e[s], a.hasOwnProperty(s) && g !== S && (g != null || S != null) && (s === "selected" ? t.selected = g && typeof g != "function" && typeof g != "symbol" : St(
            t,
            l,
            s,
            g,
            a,
            S
          ));
        return;
      case "img":
      case "link":
      case "area":
      case "base":
      case "br":
      case "col":
      case "embed":
      case "hr":
      case "keygen":
      case "meta":
      case "param":
      case "source":
      case "track":
      case "wbr":
      case "menuitem":
        for (var K in e)
          g = e[K], e.hasOwnProperty(K) && g != null && !a.hasOwnProperty(K) && St(t, l, K, null, a, g);
        for (m in a)
          if (g = a[m], S = e[m], a.hasOwnProperty(m) && g !== S && (g != null || S != null))
            switch (m) {
              case "children":
              case "dangerouslySetInnerHTML":
                if (g != null)
                  throw Error(o(137, l));
                break;
              default:
                St(
                  t,
                  l,
                  m,
                  g,
                  a,
                  S
                );
            }
        return;
      default:
        if (Ai(l)) {
          for (var bt in e)
            g = e[bt], e.hasOwnProperty(bt) && g !== void 0 && !a.hasOwnProperty(bt) && sf(
              t,
              l,
              bt,
              void 0,
              a,
              g
            );
          for (j in a)
            g = a[j], S = e[j], !a.hasOwnProperty(j) || g === S || g === void 0 && S === void 0 || sf(
              t,
              l,
              j,
              g,
              a,
              S
            );
          return;
        }
    }
    for (var y in e)
      g = e[y], e.hasOwnProperty(y) && g != null && !a.hasOwnProperty(y) && St(t, l, y, null, a, g);
    for (z in a)
      g = a[z], S = e[z], !a.hasOwnProperty(z) || g === S || g == null && S == null || St(t, l, z, g, a, S);
  }
  var rf = null, of = null;
  function Fn(t) {
    return t.nodeType === 9 ? t : t.ownerDocument;
  }
  function Fo(t) {
    switch (t) {
      case "http://www.w3.org/2000/svg":
        return 1;
      case "http://www.w3.org/1998/Math/MathML":
        return 2;
      default:
        return 0;
    }
  }
  function Po(t, l) {
    if (t === 0)
      switch (l) {
        case "svg":
          return 1;
        case "math":
          return 2;
        default:
          return 0;
      }
    return t === 1 && l === "foreignObject" ? 0 : t;
  }
  function df(t, l) {
    return t === "textarea" || t === "noscript" || typeof l.children == "string" || typeof l.children == "number" || typeof l.children == "bigint" || typeof l.dangerouslySetInnerHTML == "object" && l.dangerouslySetInnerHTML !== null && l.dangerouslySetInnerHTML.__html != null;
  }
  var hf = null;
  function T0() {
    var t = window.event;
    return t && t.type === "popstate" ? t === hf ? !1 : (hf = t, !0) : (hf = null, !1);
  }
  var Io = typeof setTimeout == "function" ? setTimeout : void 0, A0 = typeof clearTimeout == "function" ? clearTimeout : void 0, td = typeof Promise == "function" ? Promise : void 0, E0 = typeof queueMicrotask == "function" ? queueMicrotask : typeof td < "u" ? function(t) {
    return td.resolve(null).then(t).catch(z0);
  } : Io;
  function z0(t) {
    setTimeout(function() {
      throw t;
    });
  }
  function qe(t) {
    return t === "head";
  }
  function ld(t, l) {
    var e = l, a = 0, u = 0;
    do {
      var n = e.nextSibling;
      if (t.removeChild(e), n && n.nodeType === 8)
        if (e = n.data, e === "/$") {
          if (0 < a && 8 > a) {
            e = a;
            var i = t.ownerDocument;
            if (e & 1 && Gu(i.documentElement), e & 2 && Gu(i.body), e & 4)
              for (e = i.head, Gu(e), i = e.firstChild; i; ) {
                var f = i.nextSibling, s = i.nodeName;
                i[tu] || s === "SCRIPT" || s === "STYLE" || s === "LINK" && i.rel.toLowerCase() === "stylesheet" || e.removeChild(i), i = f;
              }
          }
          if (u === 0) {
            t.removeChild(n), Ju(l);
            return;
          }
          u--;
        } else
          e === "$" || e === "$?" || e === "$!" ? u++ : a = e.charCodeAt(0) - 48;
      else a = 0;
      e = n;
    } while (e);
    Ju(l);
  }
  function yf(t) {
    var l = t.firstChild;
    for (l && l.nodeType === 10 && (l = l.nextSibling); l; ) {
      var e = l;
      switch (l = l.nextSibling, e.nodeName) {
        case "HTML":
        case "HEAD":
        case "BODY":
          yf(e), Si(e);
          continue;
        case "SCRIPT":
        case "STYLE":
          continue;
        case "LINK":
          if (e.rel.toLowerCase() === "stylesheet") continue;
      }
      t.removeChild(e);
    }
  }
  function M0(t, l, e, a) {
    for (; t.nodeType === 1; ) {
      var u = e;
      if (t.nodeName.toLowerCase() !== l.toLowerCase()) {
        if (!a && (t.nodeName !== "INPUT" || t.type !== "hidden"))
          break;
      } else if (a) {
        if (!t[tu])
          switch (l) {
            case "meta":
              if (!t.hasAttribute("itemprop")) break;
              return t;
            case "link":
              if (n = t.getAttribute("rel"), n === "stylesheet" && t.hasAttribute("data-precedence"))
                break;
              if (n !== u.rel || t.getAttribute("href") !== (u.href == null || u.href === "" ? null : u.href) || t.getAttribute("crossorigin") !== (u.crossOrigin == null ? null : u.crossOrigin) || t.getAttribute("title") !== (u.title == null ? null : u.title))
                break;
              return t;
            case "style":
              if (t.hasAttribute("data-precedence")) break;
              return t;
            case "script":
              if (n = t.getAttribute("src"), (n !== (u.src == null ? null : u.src) || t.getAttribute("type") !== (u.type == null ? null : u.type) || t.getAttribute("crossorigin") !== (u.crossOrigin == null ? null : u.crossOrigin)) && n && t.hasAttribute("async") && !t.hasAttribute("itemprop"))
                break;
              return t;
            default:
              return t;
          }
      } else if (l === "input" && t.type === "hidden") {
        var n = u.name == null ? null : "" + u.name;
        if (u.type === "hidden" && t.getAttribute("name") === n)
          return t;
      } else return t;
      if (t = Xl(t.nextSibling), t === null) break;
    }
    return null;
  }
  function N0(t, l, e) {
    if (l === "") return null;
    for (; t.nodeType !== 3; )
      if ((t.nodeType !== 1 || t.nodeName !== "INPUT" || t.type !== "hidden") && !e || (t = Xl(t.nextSibling), t === null)) return null;
    return t;
  }
  function vf(t) {
    return t.data === "$!" || t.data === "$?" && t.ownerDocument.readyState === "complete";
  }
  function D0(t, l) {
    var e = t.ownerDocument;
    if (t.data !== "$?" || e.readyState === "complete")
      l();
    else {
      var a = function() {
        l(), e.removeEventListener("DOMContentLoaded", a);
      };
      e.addEventListener("DOMContentLoaded", a), t._reactRetry = a;
    }
  }
  function Xl(t) {
    for (; t != null; t = t.nextSibling) {
      var l = t.nodeType;
      if (l === 1 || l === 3) break;
      if (l === 8) {
        if (l = t.data, l === "$" || l === "$!" || l === "$?" || l === "F!" || l === "F")
          break;
        if (l === "/$") return null;
      }
    }
    return t;
  }
  var mf = null;
  function ed(t) {
    t = t.previousSibling;
    for (var l = 0; t; ) {
      if (t.nodeType === 8) {
        var e = t.data;
        if (e === "$" || e === "$!" || e === "$?") {
          if (l === 0) return t;
          l--;
        } else e === "/$" && l++;
      }
      t = t.previousSibling;
    }
    return null;
  }
  function ad(t, l, e) {
    switch (l = Fn(e), t) {
      case "html":
        if (t = l.documentElement, !t) throw Error(o(452));
        return t;
      case "head":
        if (t = l.head, !t) throw Error(o(453));
        return t;
      case "body":
        if (t = l.body, !t) throw Error(o(454));
        return t;
      default:
        throw Error(o(451));
    }
  }
  function Gu(t) {
    for (var l = t.attributes; l.length; )
      t.removeAttributeNode(l[0]);
    Si(t);
  }
  var ql = /* @__PURE__ */ new Map(), ud = /* @__PURE__ */ new Set();
  function Pn(t) {
    return typeof t.getRootNode == "function" ? t.getRootNode() : t.nodeType === 9 ? t : t.ownerDocument;
  }
  var me = H.d;
  H.d = {
    f: O0,
    r: R0,
    D: _0,
    C: U0,
    L: C0,
    m: H0,
    X: B0,
    S: q0,
    M: Y0
  };
  function O0() {
    var t = me.f(), l = Ln();
    return t || l;
  }
  function R0(t) {
    var l = va(t);
    l !== null && l.tag === 5 && l.type === "form" ? Tr(l) : me.r(t);
  }
  var $a = typeof document > "u" ? null : document;
  function nd(t, l, e) {
    var a = $a;
    if (a && typeof l == "string" && l) {
      var u = Dl(l);
      u = 'link[rel="' + t + '"][href="' + u + '"]', typeof e == "string" && (u += '[crossorigin="' + e + '"]'), ud.has(u) || (ud.add(u), t = { rel: t, crossOrigin: e, href: l }, a.querySelector(u) === null && (l = a.createElement("link"), tl(l, "link", t), wt(l), a.head.appendChild(l)));
    }
  }
  function _0(t) {
    me.D(t), nd("dns-prefetch", t, null);
  }
  function U0(t, l) {
    me.C(t, l), nd("preconnect", t, l);
  }
  function C0(t, l, e) {
    me.L(t, l, e);
    var a = $a;
    if (a && t && l) {
      var u = 'link[rel="preload"][as="' + Dl(l) + '"]';
      l === "image" && e && e.imageSrcSet ? (u += '[imagesrcset="' + Dl(
        e.imageSrcSet
      ) + '"]', typeof e.imageSizes == "string" && (u += '[imagesizes="' + Dl(
        e.imageSizes
      ) + '"]')) : u += '[href="' + Dl(t) + '"]';
      var n = u;
      switch (l) {
        case "style":
          n = Wa(t);
          break;
        case "script":
          n = Fa(t);
      }
      ql.has(n) || (t = C(
        {
          rel: "preload",
          href: l === "image" && e && e.imageSrcSet ? void 0 : t,
          as: l
        },
        e
      ), ql.set(n, t), a.querySelector(u) !== null || l === "style" && a.querySelector(Xu(n)) || l === "script" && a.querySelector(Qu(n)) || (l = a.createElement("link"), tl(l, "link", t), wt(l), a.head.appendChild(l)));
    }
  }
  function H0(t, l) {
    me.m(t, l);
    var e = $a;
    if (e && t) {
      var a = l && typeof l.as == "string" ? l.as : "script", u = 'link[rel="modulepreload"][as="' + Dl(a) + '"][href="' + Dl(t) + '"]', n = u;
      switch (a) {
        case "audioworklet":
        case "paintworklet":
        case "serviceworker":
        case "sharedworker":
        case "worker":
        case "script":
          n = Fa(t);
      }
      if (!ql.has(n) && (t = C({ rel: "modulepreload", href: t }, l), ql.set(n, t), e.querySelector(u) === null)) {
        switch (a) {
          case "audioworklet":
          case "paintworklet":
          case "serviceworker":
          case "sharedworker":
          case "worker":
          case "script":
            if (e.querySelector(Qu(n)))
              return;
        }
        a = e.createElement("link"), tl(a, "link", t), wt(a), e.head.appendChild(a);
      }
    }
  }
  function q0(t, l, e) {
    me.S(t, l, e);
    var a = $a;
    if (a && t) {
      var u = ma(a).hoistableStyles, n = Wa(t);
      l = l || "default";
      var i = u.get(n);
      if (!i) {
        var f = { loading: 0, preload: null };
        if (i = a.querySelector(
          Xu(n)
        ))
          f.loading = 5;
        else {
          t = C(
            { rel: "stylesheet", href: t, "data-precedence": l },
            e
          ), (e = ql.get(n)) && gf(t, e);
          var s = i = a.createElement("link");
          wt(s), tl(s, "link", t), s._p = new Promise(function(m, j) {
            s.onload = m, s.onerror = j;
          }), s.addEventListener("load", function() {
            f.loading |= 1;
          }), s.addEventListener("error", function() {
            f.loading |= 2;
          }), f.loading |= 4, In(i, l, a);
        }
        i = {
          type: "stylesheet",
          instance: i,
          count: 1,
          state: f
        }, u.set(n, i);
      }
    }
  }
  function B0(t, l) {
    me.X(t, l);
    var e = $a;
    if (e && t) {
      var a = ma(e).hoistableScripts, u = Fa(t), n = a.get(u);
      n || (n = e.querySelector(Qu(u)), n || (t = C({ src: t, async: !0 }, l), (l = ql.get(u)) && Sf(t, l), n = e.createElement("script"), wt(n), tl(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function Y0(t, l) {
    me.M(t, l);
    var e = $a;
    if (e && t) {
      var a = ma(e).hoistableScripts, u = Fa(t), n = a.get(u);
      n || (n = e.querySelector(Qu(u)), n || (t = C({ src: t, async: !0, type: "module" }, l), (l = ql.get(u)) && Sf(t, l), n = e.createElement("script"), wt(n), tl(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function id(t, l, e, a) {
    var u = (u = $.current) ? Pn(u) : null;
    if (!u) throw Error(o(446));
    switch (t) {
      case "meta":
      case "title":
        return null;
      case "style":
        return typeof e.precedence == "string" && typeof e.href == "string" ? (l = Wa(e.href), e = ma(
          u
        ).hoistableStyles, a = e.get(l), a || (a = {
          type: "style",
          instance: null,
          count: 0,
          state: null
        }, e.set(l, a)), a) : { type: "void", instance: null, count: 0, state: null };
      case "link":
        if (e.rel === "stylesheet" && typeof e.href == "string" && typeof e.precedence == "string") {
          t = Wa(e.href);
          var n = ma(
            u
          ).hoistableStyles, i = n.get(t);
          if (i || (u = u.ownerDocument || u, i = {
            type: "stylesheet",
            instance: null,
            count: 0,
            state: { loading: 0, preload: null }
          }, n.set(t, i), (n = u.querySelector(
            Xu(t)
          )) && !n._p && (i.instance = n, i.state.loading = 5), ql.has(t) || (e = {
            rel: "preload",
            as: "style",
            href: e.href,
            crossOrigin: e.crossOrigin,
            integrity: e.integrity,
            media: e.media,
            hrefLang: e.hrefLang,
            referrerPolicy: e.referrerPolicy
          }, ql.set(t, e), n || G0(
            u,
            t,
            e,
            i.state
          ))), l && a === null)
            throw Error(o(528, ""));
          return i;
        }
        if (l && a !== null)
          throw Error(o(529, ""));
        return null;
      case "script":
        return l = e.async, e = e.src, typeof e == "string" && l && typeof l != "function" && typeof l != "symbol" ? (l = Fa(e), e = ma(
          u
        ).hoistableScripts, a = e.get(l), a || (a = {
          type: "script",
          instance: null,
          count: 0,
          state: null
        }, e.set(l, a)), a) : { type: "void", instance: null, count: 0, state: null };
      default:
        throw Error(o(444, t));
    }
  }
  function Wa(t) {
    return 'href="' + Dl(t) + '"';
  }
  function Xu(t) {
    return 'link[rel="stylesheet"][' + t + "]";
  }
  function cd(t) {
    return C({}, t, {
      "data-precedence": t.precedence,
      precedence: null
    });
  }
  function G0(t, l, e, a) {
    t.querySelector('link[rel="preload"][as="style"][' + l + "]") ? a.loading = 1 : (l = t.createElement("link"), a.preload = l, l.addEventListener("load", function() {
      return a.loading |= 1;
    }), l.addEventListener("error", function() {
      return a.loading |= 2;
    }), tl(l, "link", e), wt(l), t.head.appendChild(l));
  }
  function Fa(t) {
    return '[src="' + Dl(t) + '"]';
  }
  function Qu(t) {
    return "script[async]" + t;
  }
  function fd(t, l, e) {
    if (l.count++, l.instance === null)
      switch (l.type) {
        case "style":
          var a = t.querySelector(
            'style[data-href~="' + Dl(e.href) + '"]'
          );
          if (a)
            return l.instance = a, wt(a), a;
          var u = C({}, e, {
            "data-href": e.href,
            "data-precedence": e.precedence,
            href: null,
            precedence: null
          });
          return a = (t.ownerDocument || t).createElement(
            "style"
          ), wt(a), tl(a, "style", u), In(a, e.precedence, t), l.instance = a;
        case "stylesheet":
          u = Wa(e.href);
          var n = t.querySelector(
            Xu(u)
          );
          if (n)
            return l.state.loading |= 4, l.instance = n, wt(n), n;
          a = cd(e), (u = ql.get(u)) && gf(a, u), n = (t.ownerDocument || t).createElement("link"), wt(n);
          var i = n;
          return i._p = new Promise(function(f, s) {
            i.onload = f, i.onerror = s;
          }), tl(n, "link", a), l.state.loading |= 4, In(n, e.precedence, t), l.instance = n;
        case "script":
          return n = Fa(e.src), (u = t.querySelector(
            Qu(n)
          )) ? (l.instance = u, wt(u), u) : (a = e, (u = ql.get(n)) && (a = C({}, e), Sf(a, u)), t = t.ownerDocument || t, u = t.createElement("script"), wt(u), tl(u, "link", a), t.head.appendChild(u), l.instance = u);
        case "void":
          return null;
        default:
          throw Error(o(443, l.type));
      }
    else
      l.type === "stylesheet" && (l.state.loading & 4) === 0 && (a = l.instance, l.state.loading |= 4, In(a, e.precedence, t));
    return l.instance;
  }
  function In(t, l, e) {
    for (var a = e.querySelectorAll(
      'link[rel="stylesheet"][data-precedence],style[data-precedence]'
    ), u = a.length ? a[a.length - 1] : null, n = u, i = 0; i < a.length; i++) {
      var f = a[i];
      if (f.dataset.precedence === l) n = f;
      else if (n !== u) break;
    }
    n ? n.parentNode.insertBefore(t, n.nextSibling) : (l = e.nodeType === 9 ? e.head : e, l.insertBefore(t, l.firstChild));
  }
  function gf(t, l) {
    t.crossOrigin == null && (t.crossOrigin = l.crossOrigin), t.referrerPolicy == null && (t.referrerPolicy = l.referrerPolicy), t.title == null && (t.title = l.title);
  }
  function Sf(t, l) {
    t.crossOrigin == null && (t.crossOrigin = l.crossOrigin), t.referrerPolicy == null && (t.referrerPolicy = l.referrerPolicy), t.integrity == null && (t.integrity = l.integrity);
  }
  var ti = null;
  function sd(t, l, e) {
    if (ti === null) {
      var a = /* @__PURE__ */ new Map(), u = ti = /* @__PURE__ */ new Map();
      u.set(e, a);
    } else
      u = ti, a = u.get(e), a || (a = /* @__PURE__ */ new Map(), u.set(e, a));
    if (a.has(t)) return a;
    for (a.set(t, null), e = e.getElementsByTagName(t), u = 0; u < e.length; u++) {
      var n = e[u];
      if (!(n[tu] || n[el] || t === "link" && n.getAttribute("rel") === "stylesheet") && n.namespaceURI !== "http://www.w3.org/2000/svg") {
        var i = n.getAttribute(l) || "";
        i = t + i;
        var f = a.get(i);
        f ? f.push(n) : a.set(i, [n]);
      }
    }
    return a;
  }
  function rd(t, l, e) {
    t = t.ownerDocument || t, t.head.insertBefore(
      e,
      l === "title" ? t.querySelector("head > title") : null
    );
  }
  function X0(t, l, e) {
    if (e === 1 || l.itemProp != null) return !1;
    switch (t) {
      case "meta":
      case "title":
        return !0;
      case "style":
        if (typeof l.precedence != "string" || typeof l.href != "string" || l.href === "")
          break;
        return !0;
      case "link":
        if (typeof l.rel != "string" || typeof l.href != "string" || l.href === "" || l.onLoad || l.onError)
          break;
        return l.rel === "stylesheet" ? (t = l.disabled, typeof l.precedence == "string" && t == null) : !0;
      case "script":
        if (l.async && typeof l.async != "function" && typeof l.async != "symbol" && !l.onLoad && !l.onError && l.src && typeof l.src == "string")
          return !0;
    }
    return !1;
  }
  function od(t) {
    return !(t.type === "stylesheet" && (t.state.loading & 3) === 0);
  }
  var Zu = null;
  function Q0() {
  }
  function Z0(t, l, e) {
    if (Zu === null) throw Error(o(475));
    var a = Zu;
    if (l.type === "stylesheet" && (typeof e.media != "string" || matchMedia(e.media).matches !== !1) && (l.state.loading & 4) === 0) {
      if (l.instance === null) {
        var u = Wa(e.href), n = t.querySelector(
          Xu(u)
        );
        if (n) {
          t = n._p, t !== null && typeof t == "object" && typeof t.then == "function" && (a.count++, a = li.bind(a), t.then(a, a)), l.state.loading |= 4, l.instance = n, wt(n);
          return;
        }
        n = t.ownerDocument || t, e = cd(e), (u = ql.get(u)) && gf(e, u), n = n.createElement("link"), wt(n);
        var i = n;
        i._p = new Promise(function(f, s) {
          i.onload = f, i.onerror = s;
        }), tl(n, "link", e), l.instance = n;
      }
      a.stylesheets === null && (a.stylesheets = /* @__PURE__ */ new Map()), a.stylesheets.set(l, t), (t = l.state.preload) && (l.state.loading & 3) === 0 && (a.count++, l = li.bind(a), t.addEventListener("load", l), t.addEventListener("error", l));
    }
  }
  function L0() {
    if (Zu === null) throw Error(o(475));
    var t = Zu;
    return t.stylesheets && t.count === 0 && bf(t, t.stylesheets), 0 < t.count ? function(l) {
      var e = setTimeout(function() {
        if (t.stylesheets && bf(t, t.stylesheets), t.unsuspend) {
          var a = t.unsuspend;
          t.unsuspend = null, a();
        }
      }, 6e4);
      return t.unsuspend = l, function() {
        t.unsuspend = null, clearTimeout(e);
      };
    } : null;
  }
  function li() {
    if (this.count--, this.count === 0) {
      if (this.stylesheets) bf(this, this.stylesheets);
      else if (this.unsuspend) {
        var t = this.unsuspend;
        this.unsuspend = null, t();
      }
    }
  }
  var ei = null;
  function bf(t, l) {
    t.stylesheets = null, t.unsuspend !== null && (t.count++, ei = /* @__PURE__ */ new Map(), l.forEach(V0, t), ei = null, li.call(t));
  }
  function V0(t, l) {
    if (!(l.state.loading & 4)) {
      var e = ei.get(t);
      if (e) var a = e.get(null);
      else {
        e = /* @__PURE__ */ new Map(), ei.set(t, e);
        for (var u = t.querySelectorAll(
          "link[data-precedence],style[data-precedence]"
        ), n = 0; n < u.length; n++) {
          var i = u[n];
          (i.nodeName === "LINK" || i.getAttribute("media") !== "not all") && (e.set(i.dataset.precedence, i), a = i);
        }
        a && e.set(null, a);
      }
      u = l.instance, i = u.getAttribute("data-precedence"), n = e.get(i) || a, n === a && e.set(null, u), e.set(i, u), this.count++, a = li.bind(this), u.addEventListener("load", a), u.addEventListener("error", a), n ? n.parentNode.insertBefore(u, n.nextSibling) : (t = t.nodeType === 9 ? t.head : t, t.insertBefore(u, t.firstChild)), l.state.loading |= 4;
    }
  }
  var Lu = {
    $$typeof: Et,
    Provider: null,
    Consumer: null,
    _currentValue: w,
    _currentValue2: w,
    _threadCount: 0
  };
  function w0(t, l, e, a, u, n, i, f) {
    this.tag = 1, this.containerInfo = t, this.pingCache = this.current = this.pendingChildren = null, this.timeoutHandle = -1, this.callbackNode = this.next = this.pendingContext = this.context = this.cancelPendingCommit = null, this.callbackPriority = 0, this.expirationTimes = F(-1), this.entangledLanes = this.shellSuspendCounter = this.errorRecoveryDisabledLanes = this.expiredLanes = this.warmLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0, this.entanglements = F(0), this.hiddenUpdates = F(null), this.identifierPrefix = a, this.onUncaughtError = u, this.onCaughtError = n, this.onRecoverableError = i, this.pooledCache = null, this.pooledCacheLanes = 0, this.formState = f, this.incompleteTransitions = /* @__PURE__ */ new Map();
  }
  function dd(t, l, e, a, u, n, i, f, s, m, j, z) {
    return t = new w0(
      t,
      l,
      e,
      i,
      f,
      s,
      m,
      z
    ), l = 1, n === !0 && (l |= 24), n = pl(3, null, null, l), t.current = n, n.stateNode = t, l = Ii(), l.refCount++, t.pooledCache = l, l.refCount++, n.memoizedState = {
      element: a,
      isDehydrated: e,
      cache: l
    }, ac(n), t;
  }
  function hd(t) {
    return t ? (t = Na, t) : Na;
  }
  function yd(t, l, e, a, u, n) {
    u = hd(u), a.context === null ? a.context = u : a.pendingContext = u, a = Te(l), a.payload = { element: e }, n = n === void 0 ? null : n, n !== null && (a.callback = n), e = Ae(t, a, l), e !== null && (zl(e, t, l), bu(e, t, l));
  }
  function vd(t, l) {
    if (t = t.memoizedState, t !== null && t.dehydrated !== null) {
      var e = t.retryLane;
      t.retryLane = e !== 0 && e < l ? e : l;
    }
  }
  function xf(t, l) {
    vd(t, l), (t = t.alternate) && vd(t, l);
  }
  function md(t) {
    if (t.tag === 13) {
      var l = Ma(t, 67108864);
      l !== null && zl(l, t, 67108864), xf(t, 67108864);
    }
  }
  var ai = !0;
  function K0(t, l, e, a) {
    var u = p.T;
    p.T = null;
    var n = H.p;
    try {
      H.p = 2, pf(t, l, e, a);
    } finally {
      H.p = n, p.T = u;
    }
  }
  function J0(t, l, e, a) {
    var u = p.T;
    p.T = null;
    var n = H.p;
    try {
      H.p = 8, pf(t, l, e, a);
    } finally {
      H.p = n, p.T = u;
    }
  }
  function pf(t, l, e, a) {
    if (ai) {
      var u = jf(a);
      if (u === null)
        ff(
          t,
          l,
          a,
          ui,
          e
        ), Sd(t, a);
      else if ($0(
        u,
        t,
        l,
        e,
        a
      ))
        a.stopPropagation();
      else if (Sd(t, a), l & 4 && -1 < k0.indexOf(t)) {
        for (; u !== null; ) {
          var n = va(u);
          if (n !== null)
            switch (n.tag) {
              case 3:
                if (n = n.stateNode, n.current.memoizedState.isDehydrated) {
                  var i = te(n.pendingLanes);
                  if (i !== 0) {
                    var f = n;
                    for (f.pendingLanes |= 2, f.entangledLanes |= 2; i; ) {
                      var s = 1 << 31 - ct(i);
                      f.entanglements[1] |= s, i &= ~s;
                    }
                    Wl(n), (vt & 6) === 0 && (Qn = gl() + 500, qu(0));
                  }
                }
                break;
              case 13:
                f = Ma(n, 2), f !== null && zl(f, n, 2), Ln(), xf(n, 2);
            }
          if (n = jf(a), n === null && ff(
            t,
            l,
            a,
            ui,
            e
          ), n === u) break;
          u = n;
        }
        u !== null && a.stopPropagation();
      } else
        ff(
          t,
          l,
          a,
          null,
          e
        );
    }
  }
  function jf(t) {
    return t = zi(t), Tf(t);
  }
  var ui = null;
  function Tf(t) {
    if (ui = null, t = ya(t), t !== null) {
      var l = D(t);
      if (l === null) t = null;
      else {
        var e = l.tag;
        if (e === 13) {
          if (t = U(l), t !== null) return t;
          t = null;
        } else if (e === 3) {
          if (l.stateNode.current.memoizedState.isDehydrated)
            return l.tag === 3 ? l.stateNode.containerInfo : null;
          t = null;
        } else l !== t && (t = null);
      }
    }
    return ui = t, null;
  }
  function gd(t) {
    switch (t) {
      case "beforetoggle":
      case "cancel":
      case "click":
      case "close":
      case "contextmenu":
      case "copy":
      case "cut":
      case "auxclick":
      case "dblclick":
      case "dragend":
      case "dragstart":
      case "drop":
      case "focusin":
      case "focusout":
      case "input":
      case "invalid":
      case "keydown":
      case "keypress":
      case "keyup":
      case "mousedown":
      case "mouseup":
      case "paste":
      case "pause":
      case "play":
      case "pointercancel":
      case "pointerdown":
      case "pointerup":
      case "ratechange":
      case "reset":
      case "resize":
      case "seeked":
      case "submit":
      case "toggle":
      case "touchcancel":
      case "touchend":
      case "touchstart":
      case "volumechange":
      case "change":
      case "selectionchange":
      case "textInput":
      case "compositionstart":
      case "compositionend":
      case "compositionupdate":
      case "beforeblur":
      case "afterblur":
      case "beforeinput":
      case "blur":
      case "fullscreenchange":
      case "focus":
      case "hashchange":
      case "popstate":
      case "select":
      case "selectstart":
        return 2;
      case "drag":
      case "dragenter":
      case "dragexit":
      case "dragleave":
      case "dragover":
      case "mousemove":
      case "mouseout":
      case "mouseover":
      case "pointermove":
      case "pointerout":
      case "pointerover":
      case "scroll":
      case "touchmove":
      case "wheel":
      case "mouseenter":
      case "mouseleave":
      case "pointerenter":
      case "pointerleave":
        return 8;
      case "message":
        switch (Sl()) {
          case Ze:
            return 2;
          case Il:
            return 8;
          case Ml:
          case bl:
            return 32;
          case be:
            return 268435456;
          default:
            return 32;
        }
      default:
        return 32;
    }
  }
  var Af = !1, Be = null, Ye = null, Ge = null, Vu = /* @__PURE__ */ new Map(), wu = /* @__PURE__ */ new Map(), Xe = [], k0 = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(
    " "
  );
  function Sd(t, l) {
    switch (t) {
      case "focusin":
      case "focusout":
        Be = null;
        break;
      case "dragenter":
      case "dragleave":
        Ye = null;
        break;
      case "mouseover":
      case "mouseout":
        Ge = null;
        break;
      case "pointerover":
      case "pointerout":
        Vu.delete(l.pointerId);
        break;
      case "gotpointercapture":
      case "lostpointercapture":
        wu.delete(l.pointerId);
    }
  }
  function Ku(t, l, e, a, u, n) {
    return t === null || t.nativeEvent !== n ? (t = {
      blockedOn: l,
      domEventName: e,
      eventSystemFlags: a,
      nativeEvent: n,
      targetContainers: [u]
    }, l !== null && (l = va(l), l !== null && md(l)), t) : (t.eventSystemFlags |= a, l = t.targetContainers, u !== null && l.indexOf(u) === -1 && l.push(u), t);
  }
  function $0(t, l, e, a, u) {
    switch (l) {
      case "focusin":
        return Be = Ku(
          Be,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "dragenter":
        return Ye = Ku(
          Ye,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "mouseover":
        return Ge = Ku(
          Ge,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "pointerover":
        var n = u.pointerId;
        return Vu.set(
          n,
          Ku(
            Vu.get(n) || null,
            t,
            l,
            e,
            a,
            u
          )
        ), !0;
      case "gotpointercapture":
        return n = u.pointerId, wu.set(
          n,
          Ku(
            wu.get(n) || null,
            t,
            l,
            e,
            a,
            u
          )
        ), !0;
    }
    return !1;
  }
  function bd(t) {
    var l = ya(t.target);
    if (l !== null) {
      var e = D(l);
      if (e !== null) {
        if (l = e.tag, l === 13) {
          if (l = U(e), l !== null) {
            t.blockedOn = l, Ia(t.priority, function() {
              if (e.tag === 13) {
                var a = El();
                a = cl(a);
                var u = Ma(e, a);
                u !== null && zl(u, e, a), xf(e, a);
              }
            });
            return;
          }
        } else if (l === 3 && e.stateNode.current.memoizedState.isDehydrated) {
          t.blockedOn = e.tag === 3 ? e.stateNode.containerInfo : null;
          return;
        }
      }
    }
    t.blockedOn = null;
  }
  function ni(t) {
    if (t.blockedOn !== null) return !1;
    for (var l = t.targetContainers; 0 < l.length; ) {
      var e = jf(t.nativeEvent);
      if (e === null) {
        e = t.nativeEvent;
        var a = new e.constructor(
          e.type,
          e
        );
        Ei = a, e.target.dispatchEvent(a), Ei = null;
      } else
        return l = va(e), l !== null && md(l), t.blockedOn = e, !1;
      l.shift();
    }
    return !0;
  }
  function xd(t, l, e) {
    ni(t) && e.delete(l);
  }
  function W0() {
    Af = !1, Be !== null && ni(Be) && (Be = null), Ye !== null && ni(Ye) && (Ye = null), Ge !== null && ni(Ge) && (Ge = null), Vu.forEach(xd), wu.forEach(xd);
  }
  function ii(t, l) {
    t.blockedOn === l && (t.blockedOn = null, Af || (Af = !0, r.unstable_scheduleCallback(
      r.unstable_NormalPriority,
      W0
    )));
  }
  var ci = null;
  function pd(t) {
    ci !== t && (ci = t, r.unstable_scheduleCallback(
      r.unstable_NormalPriority,
      function() {
        ci === t && (ci = null);
        for (var l = 0; l < t.length; l += 3) {
          var e = t[l], a = t[l + 1], u = t[l + 2];
          if (typeof a != "function") {
            if (Tf(a || e) === null)
              continue;
            break;
          }
          var n = va(e);
          n !== null && (t.splice(l, 3), l -= 3, jc(
            n,
            {
              pending: !0,
              data: u,
              method: e.method,
              action: a
            },
            a,
            u
          ));
        }
      }
    ));
  }
  function Ju(t) {
    function l(s) {
      return ii(s, t);
    }
    Be !== null && ii(Be, t), Ye !== null && ii(Ye, t), Ge !== null && ii(Ge, t), Vu.forEach(l), wu.forEach(l);
    for (var e = 0; e < Xe.length; e++) {
      var a = Xe[e];
      a.blockedOn === t && (a.blockedOn = null);
    }
    for (; 0 < Xe.length && (e = Xe[0], e.blockedOn === null); )
      bd(e), e.blockedOn === null && Xe.shift();
    if (e = (t.ownerDocument || t).$$reactFormReplay, e != null)
      for (a = 0; a < e.length; a += 3) {
        var u = e[a], n = e[a + 1], i = u[sl] || null;
        if (typeof n == "function")
          i || pd(e);
        else if (i) {
          var f = null;
          if (n && n.hasAttribute("formAction")) {
            if (u = n, i = n[sl] || null)
              f = i.formAction;
            else if (Tf(u) !== null) continue;
          } else f = i.action;
          typeof f == "function" ? e[a + 1] = f : (e.splice(a, 3), a -= 3), pd(e);
        }
      }
  }
  function Ef(t) {
    this._internalRoot = t;
  }
  fi.prototype.render = Ef.prototype.render = function(t) {
    var l = this._internalRoot;
    if (l === null) throw Error(o(409));
    var e = l.current, a = El();
    yd(e, a, t, l, null, null);
  }, fi.prototype.unmount = Ef.prototype.unmount = function() {
    var t = this._internalRoot;
    if (t !== null) {
      this._internalRoot = null;
      var l = t.containerInfo;
      yd(t.current, 2, null, t, null, null), Ln(), l[ha] = null;
    }
  };
  function fi(t) {
    this._internalRoot = t;
  }
  fi.prototype.unstable_scheduleHydration = function(t) {
    if (t) {
      var l = le();
      t = { blockedOn: null, target: t, priority: l };
      for (var e = 0; e < Xe.length && l !== 0 && l < Xe[e].priority; e++) ;
      Xe.splice(e, 0, t), e === 0 && bd(t);
    }
  };
  var jd = x.version;
  if (jd !== "19.1.1")
    throw Error(
      o(
        527,
        jd,
        "19.1.1"
      )
    );
  H.findDOMNode = function(t) {
    var l = t._reactInternals;
    if (l === void 0)
      throw typeof t.render == "function" ? Error(o(188)) : (t = Object.keys(t).join(","), Error(o(268, t)));
    return t = N(l), t = t !== null ? b(t) : null, t = t === null ? null : t.stateNode, t;
  };
  var F0 = {
    bundleType: 0,
    version: "19.1.1",
    rendererPackageName: "react-dom",
    currentDispatcherRef: p,
    reconcilerVersion: "19.1.1"
  };
  if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u") {
    var si = __REACT_DEVTOOLS_GLOBAL_HOOK__;
    if (!si.isDisabled && si.supportsFiber)
      try {
        Lt = si.inject(
          F0
        ), Ft = si;
      } catch {
      }
  }
  return $u.createRoot = function(t, l) {
    if (!_(t)) throw Error(o(299));
    var e = !1, a = "", u = Br, n = Yr, i = Gr, f = null;
    return l != null && (l.unstable_strictMode === !0 && (e = !0), l.identifierPrefix !== void 0 && (a = l.identifierPrefix), l.onUncaughtError !== void 0 && (u = l.onUncaughtError), l.onCaughtError !== void 0 && (n = l.onCaughtError), l.onRecoverableError !== void 0 && (i = l.onRecoverableError), l.unstable_transitionCallbacks !== void 0 && (f = l.unstable_transitionCallbacks)), l = dd(
      t,
      1,
      !1,
      null,
      null,
      e,
      a,
      u,
      n,
      i,
      f,
      null
    ), t[ha] = l.current, cf(t), new Ef(l);
  }, $u.hydrateRoot = function(t, l, e) {
    if (!_(t)) throw Error(o(299));
    var a = !1, u = "", n = Br, i = Yr, f = Gr, s = null, m = null;
    return e != null && (e.unstable_strictMode === !0 && (a = !0), e.identifierPrefix !== void 0 && (u = e.identifierPrefix), e.onUncaughtError !== void 0 && (n = e.onUncaughtError), e.onCaughtError !== void 0 && (i = e.onCaughtError), e.onRecoverableError !== void 0 && (f = e.onRecoverableError), e.unstable_transitionCallbacks !== void 0 && (s = e.unstable_transitionCallbacks), e.formState !== void 0 && (m = e.formState)), l = dd(
      t,
      1,
      !0,
      l,
      e ?? null,
      a,
      u,
      n,
      i,
      f,
      s,
      m
    ), l.context = hd(null), e = l.current, a = El(), a = cl(a), u = Te(a), u.callback = null, Ae(e, u, a), e = a, l.current.lanes = e, tt(l, e), Wl(l), t[ha] = l.current, cf(t), new fi(l);
  }, $u.version = "19.1.1", $u;
}
var _d;
function c1() {
  if (_d) return Mf.exports;
  _d = 1;
  function r() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(r);
      } catch (x) {
        console.error(x);
      }
  }
  return r(), Mf.exports = i1(), Mf.exports;
}
var f1 = c1(), R = Uf();
const s1 = (r) => r.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase(), Bd = (...r) => r.filter((x, A, o) => !!x && x.trim() !== "" && o.indexOf(x) === A).join(" ").trim();
var r1 = {
  xmlns: "http://www.w3.org/2000/svg",
  width: 24,
  height: 24,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 2,
  strokeLinecap: "round",
  strokeLinejoin: "round"
};
const o1 = R.forwardRef(
  ({
    color: r = "currentColor",
    size: x = 24,
    strokeWidth: A = 2,
    absoluteStrokeWidth: o,
    className: _ = "",
    children: D,
    iconNode: U,
    ...O
  }, N) => R.createElement(
    "svg",
    {
      ref: N,
      ...r1,
      width: x,
      height: x,
      stroke: r,
      strokeWidth: o ? Number(A) * 24 / Number(x) : A,
      className: Bd("lucide", _),
      ...O
    },
    [
      ...U.map(([b, C]) => R.createElement(b, C)),
      ...Array.isArray(D) ? D : [D]
    ]
  )
);
const jt = (r, x) => {
  const A = R.forwardRef(
    ({ className: o, ..._ }, D) => R.createElement(o1, {
      ref: D,
      iconNode: x,
      className: Bd(`lucide-${s1(r)}`, o),
      ..._
    })
  );
  return A.displayName = `${r}`, A;
};
const _f = jt("AudioLines", [
  ["path", { d: "M2 10v3", key: "1fnikh" }],
  ["path", { d: "M6 6v11", key: "11sgs0" }],
  ["path", { d: "M10 3v18", key: "yhl04a" }],
  ["path", { d: "M14 8v7", key: "3a1oy3" }],
  ["path", { d: "M18 5v13", key: "123xd1" }],
  ["path", { d: "M22 10v3", key: "154ddg" }]
]);
const Wu = jt("Check", [["path", { d: "M20 6 9 17l-5-5", key: "1gmf2c" }]]);
const di = jt("ChevronDown", [
  ["path", { d: "m6 9 6 6 6-6", key: "qrunsl" }]
]);
const Cf = jt("ChevronRight", [
  ["path", { d: "m9 18 6-6-6-6", key: "mthhwq" }]
]);
const d1 = jt("ChevronUp", [["path", { d: "m18 15-6-6-6 6", key: "153udz" }]]);
const Yd = jt("CircleCheck", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["path", { d: "m9 12 2 2 4-4", key: "dzmm74" }]
]);
const h1 = jt("Circle", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }]
]);
const y1 = jt("Clock3", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["polyline", { points: "12 6 12 12 16.5 12", key: "1aq6pp" }]
]);
const v1 = jt("CodeXml", [
  ["path", { d: "m18 16 4-4-4-4", key: "1inbqp" }],
  ["path", { d: "m6 8-4 4 4 4", key: "15zrgr" }],
  ["path", { d: "m14.5 4-5 16", key: "e7oirm" }]
]);
const m1 = jt("Ellipsis", [
  ["circle", { cx: "12", cy: "12", r: "1", key: "41hilf" }],
  ["circle", { cx: "19", cy: "12", r: "1", key: "1wjl8i" }],
  ["circle", { cx: "5", cy: "12", r: "1", key: "1pcz8c" }]
]);
const Hf = jt("FileAudio", [
  ["path", { d: "M17.5 22h.5a2 2 0 0 0 2-2V7l-5-5H6a2 2 0 0 0-2 2v3", key: "rslqgf" }],
  ["path", { d: "M14 2v4a2 2 0 0 0 2 2h4", key: "tnqrlb" }],
  [
    "path",
    {
      d: "M2 19a2 2 0 1 1 4 0v1a2 2 0 1 1-4 0v-4a6 6 0 0 1 12 0v4a2 2 0 1 1-4 0v-1a2 2 0 1 1 4 0",
      key: "9f7x3i"
    }
  ]
]);
const qf = jt("FolderGit2", [
  [
    "path",
    {
      d: "M9 20H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H20a2 2 0 0 1 2 2v5",
      key: "1w6njk"
    }
  ],
  ["circle", { cx: "13", cy: "12", r: "2", key: "1j92g6" }],
  ["path", { d: "M18 19c-2.8 0-5-2.2-5-5v8", key: "pkpw2h" }],
  ["circle", { cx: "20", cy: "19", r: "2", key: "1obnsp" }]
]);
const ge = jt("LoaderCircle", [
  ["path", { d: "M21 12a9 9 0 1 1-6.219-8.56", key: "13zald" }]
]);
const g1 = jt("Mic", [
  ["path", { d: "M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z", key: "131961" }],
  ["path", { d: "M19 10v2a7 7 0 0 1-14 0v-2", key: "1vc78b" }],
  ["line", { x1: "12", x2: "12", y1: "19", y2: "22", key: "x3vr5v" }]
]);
const S1 = jt("Pause", [
  ["rect", { x: "14", y: "4", width: "4", height: "16", rx: "1", key: "zuxfzm" }],
  ["rect", { x: "6", y: "4", width: "4", height: "16", rx: "1", key: "1okwgv" }]
]);
const ri = jt("PencilLine", [
  ["path", { d: "M12 20h9", key: "t2du7b" }],
  [
    "path",
    {
      d: "M16.376 3.622a1 1 0 0 1 3.002 3.002L7.368 18.635a2 2 0 0 1-.855.506l-2.872.838a.5.5 0 0 1-.62-.62l.838-2.872a2 2 0 0 1 .506-.854z",
      key: "1ykcvy"
    }
  ],
  ["path", { d: "m15 5 3 3", key: "1w25hb" }]
]);
const b1 = jt("Play", [
  ["polygon", { points: "6 3 20 12 6 21 6 3", key: "1oa8hb" }]
]);
const oi = jt("Plus", [
  ["path", { d: "M5 12h14", key: "1ays0h" }],
  ["path", { d: "M12 5v14", key: "s699le" }]
]);
const Gd = jt("Radio", [
  ["path", { d: "M4.9 19.1C1 15.2 1 8.8 4.9 4.9", key: "1vaf9d" }],
  ["path", { d: "M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5", key: "u1ii0m" }],
  ["circle", { cx: "12", cy: "12", r: "2", key: "1c9p78" }],
  ["path", { d: "M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5", key: "1j5fej" }],
  ["path", { d: "M19.1 4.9C23 8.8 23 15.1 19.1 19", key: "10b0cb" }]
]);
const x1 = jt("RefreshCw", [
  ["path", { d: "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8", key: "v9h5vc" }],
  ["path", { d: "M21 3v5h-5", key: "1q7to0" }],
  ["path", { d: "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16", key: "3uifl3" }],
  ["path", { d: "M8 16H3v5", key: "1cv678" }]
]);
const p1 = jt("Search", [
  ["circle", { cx: "11", cy: "11", r: "8", key: "4ej97u" }],
  ["path", { d: "m21 21-4.3-4.3", key: "1qie3q" }]
]);
const j1 = jt("Send", [
  [
    "path",
    {
      d: "M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z",
      key: "1ffxy3"
    }
  ],
  ["path", { d: "m21.854 2.147-10.94 10.939", key: "12cjpa" }]
]);
const T1 = jt("Settings2", [
  ["path", { d: "M20 7h-9", key: "3s1dr2" }],
  ["path", { d: "M14 17H5", key: "gfn3mx" }],
  ["circle", { cx: "17", cy: "17", r: "3", key: "18b49y" }],
  ["circle", { cx: "7", cy: "7", r: "3", key: "dfmy0x" }]
]);
const Ud = jt("Sparkles", [
  [
    "path",
    {
      d: "M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z",
      key: "4pj2yx"
    }
  ],
  ["path", { d: "M20 3v4", key: "1olli1" }],
  ["path", { d: "M22 5h-4", key: "1gvqau" }],
  ["path", { d: "M4 17v2", key: "vumght" }],
  ["path", { d: "M5 18H3", key: "zchphs" }]
]);
const A1 = jt("Trash2", [
  ["path", { d: "M3 6h18", key: "d0wm0j" }],
  ["path", { d: "M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6", key: "4alrt4" }],
  ["path", { d: "M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2", key: "v07s0e" }],
  ["line", { x1: "10", x2: "10", y1: "11", y2: "17", key: "1uufr5" }],
  ["line", { x1: "14", x2: "14", y1: "11", y2: "17", key: "xtxkd" }]
]);
const E1 = jt("X", [
  ["path", { d: "M18 6 6 18", key: "1bl5f8" }],
  ["path", { d: "m6 6 12 12", key: "d8bk6v" }]
]);
const z1 = jt("Zap", [
  [
    "path",
    {
      d: "M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z",
      key: "1xq2db"
    }
  ]
]);
let hi = async () => {
  throw Error("请从已连接的设备打开 SaySo");
};
function M1(r) {
  hi = r;
}
function N1(r, x, A) {
  return hi(r, x, A);
}
function D1(r) {
  return Uint8Array.from(atob(r), (x) => x.charCodeAt(0));
}
let Cd = Promise.resolve();
function O1(r, x, A) {
  const o = hi, _ = Cd.catch(() => {
  }).then(async () => {
    const { id: D } = await o("POST", "/api/audio/begin", {
      ...A,
      sessionId: r,
      sizeBytes: x.size,
      mimeType: x.type || "audio/webm"
    });
    try {
      for (let U = 0; U < x.size; U += 16e3) {
        const O = new Uint8Array(
          await x.slice(U, U + 16e3).arrayBuffer()
        );
        await o("POST", `/api/audio/${D}/chunk`, {
          offset: U,
          data: btoa(String.fromCharCode(...O))
        });
      }
      return await o("POST", `/api/audio/${D}/finish`);
    } catch (U) {
      throw await o("POST", `/api/audio/${D}/abort`).catch(
        () => {
        }
      ), U;
    }
  });
  return Cd = _, _;
}
async function R1(r, x) {
  const A = hi, o = [];
  let _ = 0, D = 0, U = "";
  do {
    if (x.aborted) throw Error("已取消");
    const O = await A("POST", `/api/audio/${r}/read`, {
      offset: _
    });
    if (!O.size || O.size > 64 * 1024 * 1024)
      throw Error("录音大小不合法");
    const N = D1(O.data);
    if (!N.length) throw Error("录音数据不完整");
    o.push(N), _ += N.length, D = O.size, U = O.mimeType;
  } while (_ < D);
  if (x.aborted) throw Error("已取消");
  return URL.createObjectURL(new Blob(o, { type: U }));
}
async function yl(r, x) {
  return N1(
    x?.method || "GET",
    r,
    x?.body ? JSON.parse(String(x.body)) : void 0
  );
}
const Qt = {
  cancelRun: (r) => yl(`/api/runs/${r}/cancel`, { method: "POST" }),
  state: () => yl("/api/state"),
  updateSettings: (r) => yl("/api/settings", {
    method: "PATCH",
    body: JSON.stringify(r)
  }),
  createProject: (r) => yl("/api/projects", {
    method: "POST",
    body: JSON.stringify(r)
  }),
  updateProject: (r, x) => yl(`/api/projects/${r}`, {
    method: "PATCH",
    body: JSON.stringify(x)
  }),
  createSession: (r, x) => yl("/api/sessions", {
    method: "POST",
    body: JSON.stringify({ projectId: r, title: x })
  }),
  updateSession: (r, x) => yl(`/api/sessions/${r}`, {
    method: "PATCH",
    body: JSON.stringify(x)
  }),
  deleteSession: (r) => yl(`/api/sessions/${r}`, { method: "DELETE" }),
  addUtterance: (r, x) => yl(`/api/sessions/${r}/utterances`, {
    method: "POST",
    body: JSON.stringify(x)
  }),
  extractActions: (r) => yl(`/api/sessions/${r}/extract-actions`, { method: "POST" }),
  createAction: (r, x) => yl(`/api/sessions/${r}/actions`, {
    method: "POST",
    body: JSON.stringify(x)
  }),
  updateAction: (r, x) => yl(`/api/actions/${r}`, {
    method: "PATCH",
    body: JSON.stringify(x)
  }),
  deleteAction: (r) => yl(`/api/actions/${r}`, { method: "DELETE" }),
  dispatchAction: (r) => yl(`/api/actions/${r}/dispatch`, { method: "POST" }),
  uploadAudio: O1
};
function Bf(r) {
  return r.trim().toLocaleLowerCase();
}
function _1(r, x) {
  const A = Bf(x);
  return A.length > 0 && r.toLocaleLowerCase().includes(A);
}
function U1(r, x) {
  const A = Bf(x);
  if (!A) return [{ text: r, match: !1 }];
  const o = r.toLocaleLowerCase(), _ = [];
  let D = 0;
  for (; D < r.length; ) {
    const U = o.indexOf(A, D);
    if (U === -1) {
      _.push({ text: r.slice(D), match: !1 });
      break;
    }
    U > D && _.push({ text: r.slice(D, U), match: !1 }), _.push({
      text: r.slice(U, U + A.length),
      match: !0
    }), D = U + A.length;
  }
  return _.length ? _ : [{ text: r, match: !1 }];
}
const C1 = 600 * 1e3, H1 = {
  analysisProvider: "codex",
  analysisModel: "default",
  analysisIntervalMinutes: 2,
  executor: "codex",
  executionModel: "default"
}, q1 = {
  projects: [],
  sessions: [],
  actions: [],
  runs: [],
  settings: H1
}, B1 = [
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini（需账号支持）" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "default", label: "Codex 默认模型（推荐）" }
], Y1 = [
  { value: "default", label: "Codex 默认模型（推荐）" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini" }
];
function Xd(r) {
  const x = Math.max(0, Math.floor(r / 1e3));
  return `${String(Math.floor(x / 60)).padStart(2, "0")}:${String(x % 60).padStart(2, "0")}`;
}
function Hd(r) {
  const x = new Date(r);
  return Number.isNaN(x.getTime()) ? r : `${x.getFullYear()}-${x.getMonth()}-${x.getDate()}-${x.getHours()}-${x.getMinutes()}`;
}
function G1(r) {
  const x = new Date(r);
  return Number.isNaN(x.getTime()) ? "--:--" : `${String(x.getHours()).padStart(2, "0")}:${String(x.getMinutes()).padStart(2, "0")}`;
}
function X1(r) {
  return r ? r < 1024 * 1024 ? `${Math.max(1, Math.round(r / 1024))} KB` : `${(r / 1024 / 1024).toFixed(1)} MB` : "大小未知";
}
function qd(r) {
  const x = new Date(r), A = /* @__PURE__ */ new Date();
  if (x.toDateString() === A.toDateString()) return "今天";
  const o = new Date(A);
  return o.setDate(A.getDate() - 1), x.toDateString() === o.toDateString() ? "昨天" : `${x.getMonth() + 1}月${x.getDate()}日`;
}
function ra({
  title: r,
  description: x,
  children: A,
  onClose: o,
  wide: _ = !1
}) {
  return /* @__PURE__ */ c.jsx(
    "div",
    {
      className: "modal-backdrop",
      onMouseDown: (D) => D.target === D.currentTarget && o(),
      children: /* @__PURE__ */ c.jsxs(
        "section",
        {
          className: `modal ${_ ? "modal-wide" : ""}`,
          role: "dialog",
          "aria-modal": "true",
          children: [
            /* @__PURE__ */ c.jsxs("div", { className: "modal-heading", children: [
              /* @__PURE__ */ c.jsxs("div", { children: [
                /* @__PURE__ */ c.jsx("h2", { children: r }),
                x && /* @__PURE__ */ c.jsx("p", { children: x })
              ] }),
              /* @__PURE__ */ c.jsx("button", { className: "icon-button", onClick: o, "aria-label": "关闭", children: /* @__PURE__ */ c.jsx(E1, { size: 18 }) })
            ] }),
            A
          ]
        }
      )
    }
  );
}
function Qd({ status: r }) {
  const x = {
    draft: "待完善",
    ready: "Ready",
    running: "Codex 工作中",
    done: "已完成",
    failed: "执行失败"
  }[r];
  return /* @__PURE__ */ c.jsxs("span", { className: `status-pill status-${r}`, children: [
    r === "running" && /* @__PURE__ */ c.jsx(ge, { size: 12 }),
    " ",
    x
  ] });
}
function Zd(r, x) {
  const [A, o] = R.useState([]), _ = r.join("\0");
  return R.useEffect(() => {
    if (!x) {
      o([]);
      return;
    }
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      o(r);
      return;
    }
    let D = !1, U = 0, O = 0, N;
    o([]);
    const b = () => {
      if (D) return;
      const C = Array.from(r[U]);
      if (O < C.length) {
        O += 1;
        const G = C.slice(0, O).join("");
        o((W) => {
          const st = W.slice();
          return st[U] = G, st;
        }), N = window.setTimeout(b, 34);
        return;
      }
      if (U < r.length - 1) {
        U += 1, O = 0, N = window.setTimeout(b, 360);
        return;
      }
      N = window.setTimeout(() => {
        U = 0, O = 0, o([]), N = window.setTimeout(b, 240);
      }, 1600);
    };
    return N = window.setTimeout(b, 160), () => {
      D = !0, N !== void 0 && window.clearTimeout(N);
    };
  }, [x, _]), A;
}
function Q1({
  onClose: r,
  deviceLabel: x,
  scope: A,
  registerClose: o
}) {
  const [_, D] = R.useState(!1), [U, O] = R.useState([]), N = R.useRef([]), b = R.useRef(/* @__PURE__ */ new Set()), C = R.useRef(async () => {
  }), [G, W] = R.useState(""), [st, B] = R.useState(!1), [rt, Ht] = R.useState(
    "discussion"
  ), [xt, qt] = R.useState(!1);
  R.useEffect(() => {
    const E = (F) => {
      (Sl.current || b.current.size) && (F.preventDefault(), F.returnValue = "");
    }, X = () => {
      C.current().catch(() => {
      }).finally(() => _t("设备连接已断开，未上传的录音可下载保存。"));
    };
    return window.addEventListener("beforeunload", E), window.addEventListener("remote-agent-disconnected", X), () => {
      Il.current = !0;
      try {
        Ze.current?.stop();
      } catch {
      }
      Ml.current?.getTracks().forEach((F) => F.stop()), bl.current && window.clearTimeout(bl.current), N.current.forEach((F) => URL.revokeObjectURL(F.url)), window.removeEventListener("beforeunload", E), window.removeEventListener("remote-agent-disconnected", X);
    };
  }, []);
  async function Et() {
    D(!0);
    try {
      return await C.current(), await Promise.allSettled([...b.current]), N.current.length && !confirm("仍有未上传录音，请先下载保存。确定离开吗？") ? !1 : (r(), !0);
    } catch (E) {
      return _t(
        E instanceof Error ? E.message : "正在保存，请稍后再试"
      ), !1;
    } finally {
      D(!1);
    }
  }
  R.useEffect(() => {
    o?.(Et);
  });
  const [nt, P] = R.useState(q1), [$t, V] = R.useState(!0), [zt, Wt] = R.useState(""), [Fl, nl] = R.useState(""), [Zt, Ql] = R.useState(
    /* @__PURE__ */ new Set()
  ), [vl, Dt] = R.useState(
    null
  ), [p, H] = R.useState(!1), [w, ot] = R.useState(null), [d, M] = R.useState(
    null
  ), [Y, q] = R.useState(null), [L, it] = R.useState(null), [$, ll] = R.useState(""), [Tt, Zl] = R.useState(""), [Pa, Se] = R.useState(""), [ml, oa] = R.useState(!1), [Pl, Fu] = R.useState(""), [yi, gl] = R.useState(0), Sl = R.useRef(null), Ze = R.useRef(
    null
  ), Il = R.useRef(!1), Ml = R.useRef(null), bl = R.useRef(null), be = R.useRef(null), Ll = R.useRef(null);
  function _t(E) {
    ll(E), window.setTimeout(() => ll(""), 3600);
  }
  const Lt = async (E = !1) => {
    try {
      const X = await Qt.state(), F = A ? {
        ...X,
        projects: X.projects.filter((tt) => tt.id === A.projectId),
        sessions: X.sessions.filter((tt) => tt.projectId === A.projectId),
        actions: X.actions.filter((tt) => tt.projectId === A.projectId),
        runs: X.runs.filter((tt) => X.actions.some((dt) => dt.id === tt.actionId && dt.projectId === A.projectId))
      } : X;
      P(F), Wt(
        (tt) => tt && F.projects.some((dt) => dt.id === tt) ? tt : F.projects[0]?.id || ""
      ), nl(
        (tt) => tt && F.sessions.some((dt) => dt.id === tt) ? tt : A?.sessionId || F.sessions.find(
          (dt) => dt.projectId === (zt || F.projects[0]?.id)
        )?.id || ""
      );
    } catch (X) {
      E || _t(X instanceof Error ? X.message : "无法连接服务");
    } finally {
      V(!1);
    }
  };
  R.useEffect(() => {
    Lt();
  }, []);
  const Ft = nt.actions.some((E) => E.status === "running") || nt.sessions.some(
    (E) => E.status === "recording" || ["queued", "running"].includes(E.analysisStatus || "")
  );
  R.useEffect(() => (Ft && !Ll.current && (Ll.current = window.setInterval(() => {
    Lt(!0);
  }, 1800)), !Ft && Ll.current && (window.clearInterval(Ll.current), Ll.current = null), () => {
    Ll.current && window.clearInterval(Ll.current), Ll.current = null;
  }), [Ft]);
  const il = nt.projects.find((E) => E.id === zt) || nt.projects[0], ct = nt.sessions.find(
    (E) => E.id === Fl
  ), Pu = nt.actions.filter((E) => E.sessionId === ct?.id).sort((E, X) => {
    const F = { running: 0, ready: 1, draft: 2, failed: 3, done: 4 };
    return F[E.status] - F[X.status] || X.createdAt.localeCompare(E.createdAt);
  }), vi = R.useMemo(
    () => new Map(
      nt.projects.map((E) => [
        E.id,
        nt.sessions.filter((X) => X.projectId === E.id)
      ])
    ),
    [nt.projects, nt.sessions]
  );
  R.useEffect(() => {
    if (!ml || !be.current) return;
    const E = () => gl(
      Date.now() - new Date(be.current.startedAt).getTime()
    );
    E();
    const X = window.setInterval(E, 1e3);
    return () => window.clearInterval(X);
  }, [ml]);
  function mi(E) {
    qt(!1), Wt(E), nl(
      nt.sessions.find((X) => X.projectId === E)?.id || ""
    ), Ql((X) => {
      const F = new Set(X);
      return F.delete(E), F;
    });
  }
  async function Le() {
    if (il)
      try {
        ml && await Vl();
        const E = await Qt.createSession(il.id, "");
        nl(E.id), qt(!1), await Lt(!0);
      } catch (E) {
        _t(E instanceof Error ? E.message : "创建会话失败");
      }
  }
  function Ve(E, X) {
    const F = Date.now(), tt = [], dt = new MediaRecorder(E, { audioBitsPerSecond: 64e3 });
    Sl.current = dt, dt.ondataavailable = (cl) => {
      cl.data.size && tt.push(cl.data);
    };
    let Vt = () => {
    };
    const Ke = new Promise((cl) => {
      Vt = cl;
    });
    dt.saved = Ke, dt.onstop = () => {
      const cl = Date.now(), Bl = new Blob(tt, {
        type: dt.mimeType || "audio/webm"
      });
      if (!Bl.size) {
        Vt();
        return;
      }
      const le = Qt.uploadAudio(X.id, Bl, {
        startedAt: new Date(F).toISOString(),
        endedAt: new Date(cl).toISOString(),
        durationMs: cl - F
      }).then(() => Lt(!0)).catch((Ia) => {
        const wl = URL.createObjectURL(Bl);
        N.current.push({
          url: wl,
          name: `sayso-${new Date(F).toISOString().replace(/:/g, "-")}.${Bl.type.includes("mp4") ? "m4a" : "webm"}`
        }), O([...N.current]), _t(`录音未上传：${Ia.message}。请下载备份。`);
      }).finally(() => {
        b.current.delete(le), Vt();
      });
      b.current.add(le);
    }, dt.start(1e3), bl.current = window.setTimeout(
      () => te(X),
      C1
    );
  }
  function te(E) {
    bl.current && window.clearTimeout(bl.current), bl.current = null;
    const X = Ml.current, F = Sl.current;
    !X || !F || F.state !== "recording" || (F.stop(), Ve(X, E));
  }
  async function da(E) {
    if (ml && await Vl(), !navigator.mediaDevices?.getUserMedia) {
      _t("当前浏览器不支持录音");
      return;
    }
    try {
      const X = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: !0, noiseSuppression: !0 }
      });
      Ml.current = X, await Qt.updateSession(E.id, { status: "recording" }), be.current = E, Il.current = !1, Ve(X, E);
      const F = window, tt = F.SpeechRecognition || F.webkitSpeechRecognition;
      if (tt) {
        const dt = new tt();
        dt.continuous = !0, dt.interimResults = !0, dt.lang = "zh-CN", dt.onresult = (Vt) => {
          let Ke = "";
          for (let cl = Vt.resultIndex; cl < Vt.results.length; cl += 1) {
            const Bl = Vt.results[cl], le = Bl[0].transcript.trim();
            Bl.isFinal && le ? Qt.addUtterance(E.id, {
              text: le,
              atMs: Date.now() - new Date(E.startedAt).getTime()
            }).then(() => Lt(!0)).catch((Ia) => _t(Ia.message)) : Ke += le;
          }
          Se(Ke);
        }, dt.onerror = (Vt) => {
          [
            "network",
            "service-not-allowed",
            "not-allowed",
            "audio-capture"
          ].includes(Vt.error) && (Il.current = !0, Se(""), _t(
            Vt.error === "network" ? "实时转写服务网络不可用；录音仍会继续并正常保存" : "录音正常；当前浏览器暂时无法使用实时转写"
          ));
        }, dt.onend = () => {
          if (!Il.current && Sl.current?.state === "recording")
            try {
              dt.start();
            } catch {
            }
        };
        try {
          dt.start();
        } catch {
          Il.current = !0, _t("录音已开始，实时转写不可用，可以打字补充。");
        }
        Ze.current = dt;
      } else _t("录音已开始；当前浏览器不提供实时转写");
      Fu(E.id), oa(!0), gl(Date.now() - new Date(E.startedAt).getTime()), await Lt(!0);
    } catch (X) {
      bl.current && window.clearTimeout(bl.current), Sl.current?.state === "recording" && Sl.current.stop(), await Qt.updateSession(E.id, { status: "paused" }).catch(() => {
      }), Ml.current?.getTracks().forEach((F) => F.stop()), Ml.current = null, Sl.current = null, _t(
        X instanceof Error ? `无法开始录音：${X.message}` : "无法开始录音"
      );
    }
  }
  async function Vl() {
    const E = be.current?.id || Pl, X = Sl.current?.saved;
    Il.current = !0;
    try {
      Ze.current?.stop();
    } catch {
    }
    Ze.current = null, bl.current && window.clearTimeout(bl.current), bl.current = null, Sl.current?.state !== "inactive" && Sl.current?.stop(), Ml.current?.getTracks().forEach((F) => F.stop()), Ml.current = null, Sl.current = null, be.current = null, Se(""), oa(!1), Fu(""), await X, await Promise.allSettled([...b.current]), E && (await Qt.updateSession(E, { status: "paused" }), await Lt(!0));
  }
  C.current = Vl;
  async function we(E, X) {
    try {
      await E(), X && _t(X), await Lt(!0);
    } catch (F) {
      _t(F instanceof Error ? F.message : "操作失败");
    }
  }
  return $t ? /* @__PURE__ */ c.jsxs("div", { className: "loading-screen", children: [
    /* @__PURE__ */ c.jsx("div", { className: "brand-mark", children: /* @__PURE__ */ c.jsx(_f, { size: 22 }) }),
    /* @__PURE__ */ c.jsx("p", { children: "正在打开 SaySo…" }),
    /* @__PURE__ */ c.jsx("button", { onClick: r, children: "返回工作台" })
  ] }) : /* @__PURE__ */ c.jsxs(
    "div",
    {
      className: `sayso-layout ${A ? "embedded" : ""} pane-${rt} ${xt ? "tree-open" : ""}`,
      children: [
        /* @__PURE__ */ c.jsxs("header", { className: "remote-header", children: [
          /* @__PURE__ */ c.jsxs("button", { onClick: () => {
            Et();
          }, disabled: _, children: [
            "← ",
            _ ? "正在保存…" : "工作台"
          ] }),
          /* @__PURE__ */ c.jsx("strong", { children: "SaySo" }),
          /* @__PURE__ */ c.jsx("span", { children: x }),
          /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "mobile-tree-button",
              onClick: () => qt(!xt),
              children: "项目 / 会话"
            }
          ),
          A && /* @__PURE__ */ c.jsx("select", { "aria-label": "SaySo 历史讨论", value: Fl, onChange: (E) => {
            ml || nl(E.target.value);
          }, disabled: ml, children: nt.sessions.map((E) => /* @__PURE__ */ c.jsx("option", { value: E.id, children: E.title }, E.id)) })
        ] }),
        U.length > 0 && /* @__PURE__ */ c.jsxs("div", { className: "audio-recovery", children: [
          "未上传录音，请下载保存：",
          U.map((E) => /* @__PURE__ */ c.jsx("a", { href: E.url, download: E.name, children: "下载录音备份" }, E.url))
        ] }),
        /* @__PURE__ */ c.jsxs("div", { className: "app-shell", children: [
          /* @__PURE__ */ c.jsxs("aside", { className: "sidebar", children: [
            /* @__PURE__ */ c.jsxs("div", { className: "brand-row", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "brand", children: [
                /* @__PURE__ */ c.jsx("div", { className: "brand-mark", children: /* @__PURE__ */ c.jsx(_f, { size: 21 }) }),
                /* @__PURE__ */ c.jsx("span", { children: "SaySo" })
              ] }),
              /* @__PURE__ */ c.jsx(
                "button",
                {
                  className: "system-settings",
                  onClick: () => H(!0),
                  "aria-label": "系统设置",
                  children: /* @__PURE__ */ c.jsx(T1, { size: 17 })
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs(
              "button",
              {
                className: "new-session-button",
                onClick: () => {
                  Le();
                },
                disabled: !il,
                children: [
                  /* @__PURE__ */ c.jsx(oi, { size: 16 }),
                  " 新会话 ",
                  /* @__PURE__ */ c.jsx("span", { children: "⌘ N" })
                ]
              }
            ),
            /* @__PURE__ */ c.jsxs("div", { className: "tree-heading", children: [
              /* @__PURE__ */ c.jsx("span", { children: "项目" }),
              /* @__PURE__ */ c.jsx(
                "button",
                {
                  onClick: () => Dt("new"),
                  "aria-label": "添加项目",
                  children: /* @__PURE__ */ c.jsx(oi, { size: 14 })
                }
              )
            ] }),
            /* @__PURE__ */ c.jsx("div", { className: "project-tree", children: nt.projects.map((E) => {
              const X = vi.get(E.id) || [], F = Zt.has(E.id);
              return /* @__PURE__ */ c.jsxs("div", { className: "project-branch", children: [
                /* @__PURE__ */ c.jsxs(
                  "div",
                  {
                    className: `project-node ${il?.id === E.id ? "selected" : ""}`,
                    children: [
                      /* @__PURE__ */ c.jsxs(
                        "button",
                        {
                          className: "project-select",
                          onClick: () => mi(E.id),
                          children: [
                            /* @__PURE__ */ c.jsx(
                              "span",
                              {
                                className: "tree-chevron",
                                onClick: (tt) => {
                                  tt.stopPropagation(), Ql((dt) => {
                                    const Vt = new Set(dt);
                                    return Vt.has(E.id) ? Vt.delete(E.id) : Vt.add(E.id), Vt;
                                  });
                                },
                                children: F ? /* @__PURE__ */ c.jsx(Cf, { size: 13 }) : /* @__PURE__ */ c.jsx(di, { size: 13 })
                              }
                            ),
                            /* @__PURE__ */ c.jsx(qf, { size: 15 }),
                            /* @__PURE__ */ c.jsx("strong", { children: E.name })
                          ]
                        }
                      ),
                      /* @__PURE__ */ c.jsx(
                        "button",
                        {
                          className: "tree-edit",
                          onClick: () => Dt(E),
                          "aria-label": `编辑项目 ${E.name}`,
                          children: /* @__PURE__ */ c.jsx(ri, { size: 13 })
                        }
                      )
                    ]
                  }
                ),
                !F && /* @__PURE__ */ c.jsxs("div", { className: "session-tree", children: [
                  X.map((tt) => {
                    const dt = ml && Pl === tt.id;
                    return /* @__PURE__ */ c.jsxs(
                      "div",
                      {
                        className: `tree-session ${ct?.id === tt.id ? "active" : ""}`,
                        children: [
                          /* @__PURE__ */ c.jsxs(
                            "button",
                            {
                              className: "session-select",
                              onClick: () => {
                                Wt(E.id), nl(tt.id), qt(!1);
                              },
                              children: [
                                /* @__PURE__ */ c.jsx(
                                  "span",
                                  {
                                    className: `tree-session-icon ${dt ? "recording" : ""}`,
                                    children: dt ? /* @__PURE__ */ c.jsx(Gd, { size: 12 }) : /* @__PURE__ */ c.jsx(Hf, { size: 12 })
                                  }
                                ),
                                /* @__PURE__ */ c.jsxs("span", { children: [
                                  /* @__PURE__ */ c.jsx("strong", { children: tt.title }),
                                  /* @__PURE__ */ c.jsxs("small", { children: [
                                    qd(tt.startedAt),
                                    " ·",
                                    " ",
                                    tt.transcript.length,
                                    " 条"
                                  ] })
                                ] })
                              ]
                            }
                          ),
                          /* @__PURE__ */ c.jsx(
                            "button",
                            {
                              className: "tree-edit",
                              onClick: () => ot(tt),
                              "aria-label": `重命名 ${tt.title}`,
                              children: /* @__PURE__ */ c.jsx(ri, { size: 12 })
                            }
                          )
                        ]
                      },
                      tt.id
                    );
                  }),
                  X.length === 0 && /* @__PURE__ */ c.jsx("div", { className: "tree-empty", children: "暂无会话" })
                ] })
              ] }, E.id);
            }) })
          ] }),
          /* @__PURE__ */ c.jsx("main", { className: "workspace", children: ct ? /* @__PURE__ */ c.jsxs(c.Fragment, { children: [
            /* @__PURE__ */ c.jsxs("header", { className: "topbar", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "title-stack", children: [
                /* @__PURE__ */ c.jsxs("div", { className: "eyebrow", children: [
                  /* @__PURE__ */ c.jsx("span", { children: il?.name }),
                  /* @__PURE__ */ c.jsx("span", { children: "/" }),
                  /* @__PURE__ */ c.jsx("span", { children: qd(ct.startedAt) })
                ] }),
                /* @__PURE__ */ c.jsxs("div", { className: "session-title-line", children: [
                  /* @__PURE__ */ c.jsx("h1", { children: ct.title }),
                  /* @__PURE__ */ c.jsx(
                    "button",
                    {
                      className: "small-icon",
                      onClick: () => ot(ct),
                      "aria-label": "编辑会话名称",
                      children: /* @__PURE__ */ c.jsx(ri, { size: 14 })
                    }
                  )
                ] })
              ] }),
              /* @__PURE__ */ c.jsxs("div", { className: "topbar-actions", children: [
                (ct.recordings?.length || ct.recordingPath) && /* @__PURE__ */ c.jsxs(
                  "button",
                  {
                    className: "ghost-button",
                    onClick: () => it(ct),
                    children: [
                      /* @__PURE__ */ c.jsx(b1, { size: 15 }),
                      " 录音片段"
                    ]
                  }
                ),
                ml && Pl === ct.id ? /* @__PURE__ */ c.jsxs("div", { className: "recording-control", children: [
                  /* @__PURE__ */ c.jsxs("span", { className: "recording-time", children: [
                    /* @__PURE__ */ c.jsx("i", {}),
                    " ",
                    Xd(yi)
                  ] }),
                  /* @__PURE__ */ c.jsxs(
                    "button",
                    {
                      className: "pause-button",
                      onClick: () => {
                        Vl().catch(
                          (E) => _t(E.message)
                        );
                      },
                      children: [
                        /* @__PURE__ */ c.jsx(S1, { size: 13, fill: "currentColor" }),
                        " 暂停"
                      ]
                    }
                  )
                ] }) : /* @__PURE__ */ c.jsxs(
                  "button",
                  {
                    className: "record-button",
                    onClick: () => {
                      da(ct);
                    },
                    children: [
                      /* @__PURE__ */ c.jsx(g1, { size: 15 }),
                      " 继续录音"
                    ]
                  }
                ),
                /* @__PURE__ */ c.jsx("button", { className: "icon-button", "aria-label": "更多", children: /* @__PURE__ */ c.jsx(m1, { size: 19 }) })
              ] })
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "mobile-tabs", children: [
              /* @__PURE__ */ c.jsx(
                "button",
                {
                  className: rt === "discussion" ? "active" : "",
                  onClick: () => Ht("discussion"),
                  children: "讨论记录"
                }
              ),
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: rt === "actions" ? "active" : "",
                  onClick: () => Ht("actions"),
                  children: [
                    "Action · ",
                    Pu.length
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "content-grid", children: [
              /* @__PURE__ */ c.jsx(
                V1,
                {
                  session: ct,
                  interimText: Pl === ct.id ? Pa : "",
                  isRecording: ml && Pl === ct.id,
                  search: Tt,
                  setSearch: Zl
                }
              ),
              /* @__PURE__ */ c.jsx(
                w1,
                {
                  session: ct,
                  actions: Pu,
                  runs: nt.runs,
                  settings: nt.settings,
                  onCreate: () => M("new"),
                  onExtract: () => we(
                    () => Qt.extractActions(ct.id),
                    "已开始重新分析讨论"
                  ),
                  onEdit: (E) => M(E),
                  onDelete: (E) => {
                    window.confirm(`删除“${E.title}”？`) && we(() => Qt.deleteAction(E.id));
                  },
                  onReady: (E) => we(
                    () => Qt.updateAction(E.id, {
                      status: E.status === "ready" ? "draft" : "ready"
                    })
                  ),
                  onDispatch: (E) => we(
                    async () => {
                      const X = await Qt.dispatchAction(E.id);
                      return X.webSessionId && await A?.onTaskCreated?.(X.webSessionId), X;
                    },
                    "已创建新的 Web 会话并排队执行"
                  ),
                  onViewRun: (E) => q(E)
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs(
              "form",
              {
                className: "manual-discussion",
                onSubmit: async (E) => {
                  if (E.preventDefault(), !(!G.trim() || st)) {
                    B(!0);
                    try {
                      await Qt.addUtterance(ct.id, {
                        text: G.trim(),
                        atMs: Date.now() - Date.parse(ct.startedAt)
                      }), W(""), await Lt(!0);
                    } catch (X) {
                      _t(
                        X instanceof Error ? X.message : "记录失败"
                      );
                    } finally {
                      B(!1);
                    }
                  }
                },
                children: [
                  /* @__PURE__ */ c.jsx(
                    "textarea",
                    {
                      "aria-label": "补充讨论",
                      placeholder: "也可以打字补充讨论…",
                      maxLength: 4e3,
                      rows: 2,
                      value: G,
                      onChange: (E) => W(E.target.value)
                    }
                  ),
                  /* @__PURE__ */ c.jsx("button", { disabled: st || !G.trim(), children: "记录" }),
                  /* @__PURE__ */ c.jsx("small", { children: "录音保存在目标电脑。实时转写由浏览器提供，可能使用其在线识别服务。" })
                ]
              }
            )
          ] }) : /* @__PURE__ */ c.jsx(
            Z1,
            {
              project: il,
              onStart: () => {
                Le();
              },
              onProject: () => Dt("new")
            }
          ) })
        ] }),
        p && /* @__PURE__ */ c.jsx(
          k1,
          {
            settings: nt.settings,
            onClose: () => H(!1),
            onSave: async (E) => {
              try {
                await Qt.updateSettings(E), H(!1), await Lt(!0), _t("系统设置已保存");
              } catch (X) {
                _t(
                  X instanceof Error ? X.message : "保存设置失败"
                );
              }
            }
          }
        ),
        vl && /* @__PURE__ */ c.jsx(
          F1,
          {
            project: vl === "new" ? void 0 : vl,
            onClose: () => Dt(null),
            onSave: async (E) => {
              try {
                const X = vl === "new" ? await Qt.createProject(E) : await Qt.updateProject(vl.id, E);
                Wt(X.id), Dt(null), await Lt(!0);
              } catch (X) {
                _t(
                  X instanceof Error ? X.message : "保存项目失败"
                );
              }
            }
          }
        ),
        w && /* @__PURE__ */ c.jsx(
          P1,
          {
            session: w,
            onClose: () => ot(null),
            onSave: async (E) => {
              try {
                await Qt.updateSession(w.id, { title: E }), ot(null), await Lt(!0);
              } catch (X) {
                _t(X instanceof Error ? X.message : "重命名失败");
              }
            }
          }
        ),
        d && ct && /* @__PURE__ */ c.jsx(
          I1,
          {
            action: d === "new" ? void 0 : d,
            onClose: () => M(null),
            onSave: async (E) => {
              await we(
                () => d === "new" ? Qt.createAction(ct.id, E) : Qt.updateAction(d.id, E)
              ), M(null);
            }
          }
        ),
        Y && /* @__PURE__ */ c.jsx(
          ty,
          {
            action: Y,
            state: nt,
            onClose: () => q(null)
          }
        ),
        L && /* @__PURE__ */ c.jsx(
          W1,
          {
            session: nt.sessions.find((E) => E.id === L.id) || L,
            onClose: () => it(null)
          }
        ),
        $ && /* @__PURE__ */ c.jsxs("div", { className: "toast", children: [
          /* @__PURE__ */ c.jsx(Yd, { size: 16 }),
          " ",
          $
        ] })
      ]
    }
  );
}
function Z1({
  project: r,
  onStart: x,
  onProject: A
}) {
  return /* @__PURE__ */ c.jsxs("div", { className: "empty-workspace", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "empty-orbit", children: [
      /* @__PURE__ */ c.jsx("span", { children: /* @__PURE__ */ c.jsx(_f, { size: 28 }) }),
      /* @__PURE__ */ c.jsx("i", {}),
      /* @__PURE__ */ c.jsx("i", {}),
      /* @__PURE__ */ c.jsx("i", {})
    ] }),
    /* @__PURE__ */ c.jsx("p", { className: "overline", children: "CONVERSATION → ACTION" }),
    /* @__PURE__ */ c.jsxs("h1", { children: [
      "让讨论自然地",
      /* @__PURE__ */ c.jsx("br", {}),
      "变成下一步。"
    ] }),
    /* @__PURE__ */ c.jsxs("p", { className: "empty-lead", children: [
      "新建一个暂停中的会话，准备好后再继续录音。",
      /* @__PURE__ */ c.jsx("br", {}),
      "SaySo 会持续转写、分析，并把明确的事交给 Codex。"
    ] }),
    /* @__PURE__ */ c.jsxs("div", { className: "empty-actions", children: [
      /* @__PURE__ */ c.jsxs("button", { className: "primary-button", onClick: x, children: [
        /* @__PURE__ */ c.jsx(oi, { size: 17 }),
        " 新建会话"
      ] }),
      !r && /* @__PURE__ */ c.jsxs("button", { className: "ghost-button", onClick: A, children: [
        /* @__PURE__ */ c.jsx(qf, { size: 16 }),
        " 添加项目"
      ] })
    ] })
  ] });
}
function L1({
  text: r,
  query: x
}) {
  return /* @__PURE__ */ c.jsx(c.Fragment, { children: U1(r, x).map(
    (A, o) => A.match ? /* @__PURE__ */ c.jsx(
      "mark",
      {
        className: "discussion-search-highlight",
        children: A.text
      },
      `${o}-${A.text}`
    ) : /* @__PURE__ */ c.jsx("span", { children: A.text }, `${o}-${A.text}`)
  ) });
}
function V1({
  session: r,
  interimText: x,
  isRecording: A,
  search: o,
  setSearch: _
}) {
  const D = R.useRef(null), U = R.useRef(/* @__PURE__ */ new Map()), O = Bf(o), N = R.useMemo(
    () => r.transcript.filter(
      (B) => _1(B.text, O)
    ),
    [r.transcript, O]
  ), b = R.useMemo(
    () => new Map(N.map((B, rt) => [B.id, rt])),
    [N]
  ), C = N.map((B) => B.id).join("|"), [G, W] = R.useState(-1);
  R.useEffect(() => {
    W(O && N.length ? 0 : -1);
  }, [O, r.id]), R.useEffect(() => {
    W(
      (B) => N.length === 0 ? -1 : Math.min(Math.max(B, 0), N.length - 1)
    );
  }, [N.length]), R.useEffect(() => {
    if (G < 0) return;
    U.current.get(N[G]?.id)?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [G, C, O]), R.useEffect(() => {
    O || D.current?.scrollIntoView({ behavior: "smooth" });
  }, [r.transcript.length, x, O]);
  const st = (B) => {
    N.length && W(
      (rt) => rt < 0 ? B === 1 ? 0 : N.length - 1 : (rt + B + N.length) % N.length
    );
  };
  return /* @__PURE__ */ c.jsxs("section", { className: "transcript-panel", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "panel-header", children: [
      /* @__PURE__ */ c.jsxs("div", { children: [
        /* @__PURE__ */ c.jsx("h2", { children: "讨论记录" }),
        /* @__PURE__ */ c.jsxs("span", { children: [
          r.transcript.length,
          " 条"
        ] })
      ] }),
      /* @__PURE__ */ c.jsxs("div", { className: "transcript-tools", children: [
        /* @__PURE__ */ c.jsxs(
          "div",
          {
            className: `discussion-search ${O ? "has-query" : ""}`,
            children: [
              /* @__PURE__ */ c.jsxs("label", { className: "compact-search", children: [
                /* @__PURE__ */ c.jsx(p1, { size: 14 }),
                /* @__PURE__ */ c.jsx(
                  "input",
                  {
                    "aria-label": "搜索讨论",
                    value: o,
                    onChange: (B) => _(B.target.value),
                    onKeyDown: (B) => {
                      B.key === "Enter" && (B.preventDefault(), st(B.shiftKey ? -1 : 1)), B.key === "Escape" && _("");
                    },
                    placeholder: "搜索讨论"
                  }
                )
              ] }),
              O && /* @__PURE__ */ c.jsxs("div", { className: "search-navigation", children: [
                /* @__PURE__ */ c.jsx("span", { children: N.length ? `${G + 1}/${N.length}` : "0/0" }),
                /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !N.length,
                    onClick: () => st(-1),
                    "aria-label": "上一个搜索结果",
                    children: /* @__PURE__ */ c.jsx(d1, { size: 13 })
                  }
                ),
                /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !N.length,
                    onClick: () => st(1),
                    "aria-label": "下一个搜索结果",
                    children: /* @__PURE__ */ c.jsx(di, { size: 13 })
                  }
                )
              ] })
            ]
          }
        ),
        /* @__PURE__ */ c.jsxs("span", { className: `live-chip ${A ? "on" : ""}`, children: [
          /* @__PURE__ */ c.jsx("i", {}),
          " ",
          A ? "录音中" : "已暂停"
        ] })
      ] })
    ] }),
    /* @__PURE__ */ c.jsx("div", { className: "transcript-scroll", children: r.transcript.length === 0 && !x ? /* @__PURE__ */ c.jsxs("div", { className: "transcript-empty", children: [
      /* @__PURE__ */ c.jsxs("div", { className: `sound-wave ${A ? "active" : ""}`, children: [
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {})
      ] }),
      /* @__PURE__ */ c.jsx("h3", { children: A ? "正在听…" : "讨论已暂停" }),
      /* @__PURE__ */ c.jsx("p", { children: A ? "识别出的讨论会自动出现在这里。" : "点击右上角“继续录音”开始记录。" })
    ] }) : /* @__PURE__ */ c.jsxs("div", { className: "utterance-list", children: [
      r.transcript.map((B, rt) => {
        const Ht = Hd(B.createdAt), xt = rt === 0 || Ht !== Hd(r.transcript[rt - 1].createdAt), qt = G1(B.createdAt), Et = b.get(B.id), nt = Et !== void 0 && Et === G;
        return /* @__PURE__ */ c.jsxs(
          "article",
          {
            className: `utterance ${Et !== void 0 ? "discussion-search-match" : ""} ${nt ? "current-search-match" : ""}`,
            ref: (P) => {
              P ? U.current.set(B.id, P) : U.current.delete(B.id);
            },
            children: [
              /* @__PURE__ */ c.jsx(
                "time",
                {
                  className: xt ? "minute-mark" : void 0,
                  dateTime: B.createdAt,
                  "aria-label": xt ? `系统时间 ${qt}` : void 0,
                  "aria-hidden": !xt || void 0,
                  children: xt ? qt : null
                }
              ),
              /* @__PURE__ */ c.jsx("p", { children: /* @__PURE__ */ c.jsx(
                L1,
                {
                  text: B.text,
                  query: O
                }
              ) })
            ]
          },
          B.id
        );
      }),
      x && /* @__PURE__ */ c.jsxs("article", { className: "utterance interim", children: [
        /* @__PURE__ */ c.jsxs("time", { children: [
          /* @__PURE__ */ c.jsx(Gd, { size: 10 }),
          " LIVE"
        ] }),
        /* @__PURE__ */ c.jsxs("p", { children: [
          x,
          /* @__PURE__ */ c.jsx("span", { className: "typing-cursor" })
        ] })
      ] }),
      /* @__PURE__ */ c.jsx("div", { ref: D })
    ] }) })
  ] });
}
function w1({
  session: r,
  actions: x,
  runs: A,
  settings: o,
  onCreate: _,
  onExtract: D,
  onEdit: U,
  onDelete: O,
  onReady: N,
  onDispatch: b,
  onViewRun: C
}) {
  const G = ["queued", "running"].includes(
    r.analysisStatus || ""
  ), [W, st] = R.useState(!0), [B, rt] = R.useState(!1), xt = r.analysisTraces?.find(
    (V) => V.id === r.analysisActiveTraceId
  ) || r.analysisTraces?.[0], qt = R.useRef(void 0);
  R.useEffect(() => {
    G && xt?.id && qt.current !== xt.id && (qt.current = xt.id, st(!0));
  }, [G, xt?.id]);
  const Et = [
    "读取最近 5 分钟的原始讨论",
    `对照 ${x.filter((V) => !["running", "done"].includes(V.status)).length} 条已有需求`,
    `${o.analysisModel === "default" ? "Codex" : o.analysisModel} 正在判断补充与修正`
  ], nt = xt && xt.events.length >= 3 ? xt.events.slice(-3) : Et, P = Zd(
    nt,
    G && W
  ), $t = [...P].reverse().find(Boolean);
  return /* @__PURE__ */ c.jsxs("aside", { className: "action-panel", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "action-heading", children: [
      /* @__PURE__ */ c.jsxs("div", { children: [
        /* @__PURE__ */ c.jsxs("span", { className: "action-kicker", children: [
          /* @__PURE__ */ c.jsx(z1, { size: 13, fill: "currentColor" }),
          " ACTIONS"
        ] }),
        /* @__PURE__ */ c.jsx("h2", { children: "接下来要做的事" })
      ] }),
      /* @__PURE__ */ c.jsx(
        "button",
        {
          className: "icon-button bordered",
          onClick: _,
          "aria-label": "添加 Action",
          children: /* @__PURE__ */ c.jsx(oi, { size: 18 })
        }
      )
    ] }),
    /* @__PURE__ */ c.jsxs(
      "div",
      {
        className: `analysis-strip ${r.analysisStatus === "failed" ? "failed" : ""} ${W ? "expanded" : ""}`,
        children: [
          G ? /* @__PURE__ */ c.jsx(ge, { size: 13, className: "spin" }) : /* @__PURE__ */ c.jsx(Ud, { size: 13 }),
          /* @__PURE__ */ c.jsxs(
            "button",
            {
              className: "analysis-status-copy",
              onClick: () => G && st((V) => !V),
              disabled: !G,
              children: [
                /* @__PURE__ */ c.jsx("span", { children: G ? W && $t || "分析中" : r.analysisStatus === "failed" ? r.analysisError || "分析失败，点击重试" : r.status === "recording" ? "每 2 分钟分析最近 5 分钟讨论" : `${x.length} 条 Action` }),
                G && (W ? /* @__PURE__ */ c.jsx(di, { size: 12 }) : /* @__PURE__ */ c.jsx(Cf, { size: 12 }))
              ]
            }
          ),
          xt && /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "analysis-details-button",
              onClick: () => rt(!0),
              children: "详情"
            }
          ),
          /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "analysis-refresh",
              onClick: D,
              title: "重新分析讨论",
              children: /* @__PURE__ */ c.jsx(x1, { size: 13 })
            }
          ),
          G && W && /* @__PURE__ */ c.jsx("div", { className: "analysis-detail-lines", children: nt.map((V, zt) => /* @__PURE__ */ c.jsxs(
            "div",
            {
              className: P[zt] === V ? "complete" : P[zt] ? "typing" : "",
              children: [
                /* @__PURE__ */ c.jsx("span", { children: zt + 1 }),
                /* @__PURE__ */ c.jsx("p", { children: P[zt] || " " })
              ]
            },
            `${zt}-${V}`
          )) })
        ]
      }
    ),
    /* @__PURE__ */ c.jsx("div", { className: "action-list", children: x.length === 0 ? /* @__PURE__ */ c.jsxs("div", { className: "action-empty-compact", children: [
      /* @__PURE__ */ c.jsx(Ud, { size: 17 }),
      /* @__PURE__ */ c.jsx("span", { children: G ? "正在整理讨论…" : "还没有明确的 Action" })
    ] }) : x.map((V) => {
      const zt = A.find((Wt) => Wt.id === V.runId);
      return /* @__PURE__ */ c.jsxs(
        "article",
        {
          className: `action-card card-${V.status}`,
          children: [
            /* @__PURE__ */ c.jsxs("div", { className: "action-card-top", children: [
              /* @__PURE__ */ c.jsx(Qd, { status: V.status }),
              /* @__PURE__ */ c.jsx("span", { className: `priority priority-${V.priority}`, children: V.priority === "high" ? "高优先级" : V.priority === "low" ? "低优先级" : "普通" }),
              /* @__PURE__ */ c.jsxs("div", { className: "card-tools", children: [
                !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsx("button", { onClick: () => U(V), "aria-label": "编辑", children: /* @__PURE__ */ c.jsx(ri, { size: 14 }) }),
                !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    onClick: () => O(V),
                    "aria-label": "删除",
                    children: /* @__PURE__ */ c.jsx(A1, { size: 14 })
                  }
                )
              ] })
            ] }),
            /* @__PURE__ */ c.jsx(
              K1,
              {
                action: V,
                trace: xt,
                active: G && !["running", "done"].includes(V.status)
              }
            ),
            /* @__PURE__ */ c.jsx("h3", { children: V.title }),
            V.detail && /* @__PURE__ */ c.jsx("p", { className: "action-detail", children: V.detail }),
            /* @__PURE__ */ c.jsxs("blockquote", { children: [
              /* @__PURE__ */ c.jsx("span", { children: "“" }),
              V.source
            ] }),
            V.status === "running" && /* @__PURE__ */ c.jsxs(
              "button",
              {
                className: "run-progress",
                onClick: () => C(V),
                children: [
                  /* @__PURE__ */ c.jsxs("span", { children: [
                    /* @__PURE__ */ c.jsx(ge, { size: 15 }),
                    " Codex 正在项目中工作"
                  ] }),
                  /* @__PURE__ */ c.jsx("small", { children: "查看输出" })
                ]
              }
            ),
            V.status === "done" && /* @__PURE__ */ c.jsxs(
              "button",
              {
                className: "run-complete",
                onClick: () => C(V),
                children: [
                  /* @__PURE__ */ c.jsxs("span", { children: [
                    /* @__PURE__ */ c.jsx(Yd, { size: 15 }),
                    " 已由 Codex 完成"
                  ] }),
                  /* @__PURE__ */ c.jsx("small", { children: "查看结果" })
                ]
              }
            ),
            V.status === "failed" && zt?.error && /* @__PURE__ */ c.jsx("p", { className: "run-error", children: zt.error }),
            !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsxs("div", { className: "action-card-footer", children: [
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: `ready-toggle ${V.status === "ready" ? "checked" : ""}`,
                  onClick: () => N(V),
                  children: [
                    /* @__PURE__ */ c.jsx("span", { children: V.status === "ready" && /* @__PURE__ */ c.jsx(Wu, { size: 11 }) }),
                    " ",
                    "讨论完善"
                  ]
                }
              ),
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: "dispatch-button",
                  disabled: V.status === "draft",
                  onClick: () => b(V),
                  children: [
                    /* @__PURE__ */ c.jsx(j1, { size: 14 }),
                    " ",
                    V.status === "failed" ? "重新交给 Codex" : "交给 Codex"
                  ]
                }
              )
            ] })
          ]
        },
        V.id
      );
    }) }),
    /* @__PURE__ */ c.jsxs("div", { className: "action-footnote", children: [
      /* @__PURE__ */ c.jsx(v1, { size: 13 }),
      /* @__PURE__ */ c.jsx("span", { children: "Codex 将在当前项目目录内执行" })
    ] }),
    B && xt && /* @__PURE__ */ c.jsx(
      J1,
      {
        trace: xt,
        onClose: () => rt(!1)
      }
    )
  ] });
}
function K1({
  action: r,
  trace: x,
  active: A
}) {
  const [o, _] = R.useState(!0), D = R.useRef(void 0);
  R.useEffect(() => {
    A && x?.id && D.current !== x.id && (D.current = x.id, _(!0));
  }, [A, x?.id]);
  const U = [
    `输入：核对${x?.contextLabel || "最近 5 分钟"}的 ${x?.utteranceCount || 0} 条讨论`,
    `判断“${r.title}”是否需要补充或修正`
  ], O = Zd(U, A && o);
  return A ? /* @__PURE__ */ c.jsxs("div", { className: `requirement-analysis ${o ? "expanded" : ""}`, children: [
    /* @__PURE__ */ c.jsxs(
      "button",
      {
        className: "requirement-analysis-toggle",
        onClick: () => _((N) => !N),
        children: [
          /* @__PURE__ */ c.jsx(ge, { size: 11, className: "spin" }),
          /* @__PURE__ */ c.jsx("span", { children: o ? "需求分析中" : "分析中" }),
          o ? /* @__PURE__ */ c.jsx(di, { size: 11 }) : /* @__PURE__ */ c.jsx(Cf, { size: 11 })
        ]
      }
    ),
    o && /* @__PURE__ */ c.jsx("div", { className: "requirement-analysis-lines", children: U.map((N, b) => /* @__PURE__ */ c.jsx(
      "p",
      {
        className: O[b] === N ? "complete" : O[b] ? "typing" : "",
        children: O[b] || " "
      },
      N
    )) })
  ] }) : null;
}
function J1({
  trace: r,
  onClose: x
}) {
  const A = R.useRef(null);
  R.useEffect(() => {
    A.current?.scrollIntoView({
      behavior: "smooth",
      block: "nearest"
    });
  }, [r.events.length, r.output]);
  const o = {
    queued: "排队中",
    running: "分析中",
    done: "已完成",
    failed: "失败"
  }[r.status];
  return /* @__PURE__ */ c.jsxs(
    ra,
    {
      wide: !0,
      title: "分析详情",
      description: `${r.model} · ${o} · ${r.utteranceCount} 条原始讨论`,
      onClose: x,
      children: [
        /* @__PURE__ */ c.jsxs("div", { className: "analysis-trace-scroll", children: [
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsxs("div", { className: "analysis-trace-heading", children: [
              /* @__PURE__ */ c.jsx("h3", { children: "实时过程" }),
              /* @__PURE__ */ c.jsxs("span", { className: `trace-status trace-${r.status}`, children: [
                r.status === "running" && /* @__PURE__ */ c.jsx(ge, { size: 11, className: "spin" }),
                o
              ] })
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "analysis-live-log", children: [
              r.events.map((_, D) => /* @__PURE__ */ c.jsxs("div", { children: [
                /* @__PURE__ */ c.jsx("span", { children: String(D + 1).padStart(2, "0") }),
                /* @__PURE__ */ c.jsx("p", { children: _ })
              ] }, `${D}-${_}`)),
              /* @__PURE__ */ c.jsx("div", { ref: A })
            ] })
          ] }),
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsx("h3", { children: "原始输入" }),
            /* @__PURE__ */ c.jsx("pre", { children: r.input || "暂无输入" })
          ] }),
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsx("h3", { children: "最终输出" }),
            /* @__PURE__ */ c.jsx("pre", { children: r.output || (r.status === "running" ? "等待模型返回最终结构化输出…" : r.error || "本轮没有输出") })
          ] })
        ] }),
        /* @__PURE__ */ c.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: x, children: "关闭" }) })
      ]
    }
  );
}
function k1({
  settings: r,
  onClose: x,
  onSave: A
}) {
  const [o, _] = R.useState(r), [D, U] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    ra,
    {
      title: "系统设置",
      description: "配置讨论如何被分析，以及 Action 由谁执行。",
      onClose: x,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), U(!0), A(o).finally(() => U(!1));
          },
          children: [
            /* @__PURE__ */ c.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ c.jsx("strong", { children: "讨论分析" }),
                /* @__PURE__ */ c.jsx("small", { children: "从转写中持续提炼决定和待办" })
              ] }),
              /* @__PURE__ */ c.jsxs(
                "select",
                {
                  value: o.analysisProvider,
                  onChange: (O) => _({
                    ...o,
                    analysisProvider: O.target.value
                  }),
                  children: [
                    /* @__PURE__ */ c.jsx("option", { value: "codex", children: "Codex" }),
                    /* @__PURE__ */ c.jsx("option", { value: "local", children: "本地规则" })
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ c.jsx("strong", { children: "分析模型" }),
                /* @__PURE__ */ c.jsx("small", { children: "小模型更快，也更节省额度" })
              ] }),
              /* @__PURE__ */ c.jsx(
                "select",
                {
                  disabled: o.analysisProvider === "local",
                  value: o.analysisModel,
                  onChange: (O) => _({ ...o, analysisModel: O.target.value }),
                  children: B1.map((O) => /* @__PURE__ */ c.jsx("option", { value: O.value, children: O.label }, O.value))
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ c.jsx("strong", { children: "持续分析" }),
                /* @__PURE__ */ c.jsx("small", { children: "录音中每 2 分钟运行一次，每次读取最近 5 分钟讨论" })
              ] }),
              /* @__PURE__ */ c.jsx("span", { className: "setting-fixed-value", children: "2 分钟 / 5 分钟窗口" })
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ c.jsx("strong", { children: "执行模型" }),
                /* @__PURE__ */ c.jsx("small", { children: "最终进入项目实现 Action 时使用" })
              ] }),
              /* @__PURE__ */ c.jsx(
                "select",
                {
                  value: o.executionModel,
                  onChange: (O) => _({ ...o, executionModel: O.target.value }),
                  children: Y1.map((O) => /* @__PURE__ */ c.jsx("option", { value: O.value, children: O.label }, O.value))
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ c.jsx("strong", { children: "Action 执行器" }),
                /* @__PURE__ */ c.jsx("small", { children: "在项目目录内完成已确认的工作" })
              ] }),
              /* @__PURE__ */ c.jsx("select", { value: "codex", disabled: !0, children: /* @__PURE__ */ c.jsx("option", { value: "codex", children: "Codex" }) })
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: x, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: D, children: [
                D ? /* @__PURE__ */ c.jsx(ge, { size: 15 }) : /* @__PURE__ */ c.jsx(Wu, { size: 15 }),
                " 保存设置"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function $1({
  recording: r,
  index: x
}) {
  const [A, o] = R.useState(""), [_, D] = R.useState(""), [U, O] = R.useState(!1), N = R.useRef(null);
  R.useEffect(
    () => () => {
      N.current?.abort();
    },
    []
  ), R.useEffect(
    () => () => {
      A && URL.revokeObjectURL(A);
    },
    [A]
  );
  const b = async () => {
    O(!0), D("");
    const B = new AbortController();
    N.current = B;
    try {
      o(await R1(r.id, B.signal));
    } catch (rt) {
      B.signal.aborted || D(rt instanceof Error ? rt.message : "无法读取录音");
    } finally {
      B.signal.aborted || O(!1);
    }
  }, [C, G] = R.useState(), W = r.durationMs ?? C, st = (B) => {
    if (Number.isFinite(B.duration) && B.duration > 0) {
      G(B.duration * 1e3), B.dataset.durationProbe === "active" && (B.currentTime = 0, delete B.dataset.durationProbe);
      return;
    }
    r.durationMs === void 0 && B.dataset.durationProbe !== "active" && (B.dataset.durationProbe = "active", B.currentTime = Number.MAX_SAFE_INTEGER);
  };
  return /* @__PURE__ */ c.jsxs("article", { className: "recording-row", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "recording-index", children: [
      /* @__PURE__ */ c.jsx(Hf, { size: 15 }),
      /* @__PURE__ */ c.jsx("span", { children: String(x + 1).padStart(2, "0") })
    ] }),
    /* @__PURE__ */ c.jsxs("div", { className: "recording-info", children: [
      /* @__PURE__ */ c.jsxs("div", { children: [
        /* @__PURE__ */ c.jsxs("strong", { children: [
          "录音片段 ",
          x + 1
        ] }),
        /* @__PURE__ */ c.jsx("span", { children: new Date(r.startedAt).toLocaleString("zh-CN", {
          month: "numeric",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit"
        }) })
      ] }),
      /* @__PURE__ */ c.jsxs("small", { children: [
        W !== void 0 ? Xd(W) : "读取时长中",
        " ·",
        " ",
        X1(r.sizeBytes)
      ] }),
      !A && /* @__PURE__ */ c.jsx(
        "button",
        {
          className: "ghost-button",
          disabled: U,
          onClick: () => {
            b();
          },
          children: U ? "正在从设备读取…" : "加载录音"
        }
      ),
      _ && /* @__PURE__ */ c.jsx("p", { role: "alert", children: _ }),
      A && /* @__PURE__ */ c.jsx(
        "audio",
        {
          controls: !0,
          preload: "metadata",
          src: A,
          onLoadedMetadata: (B) => st(B.currentTarget),
          onDurationChange: (B) => st(B.currentTarget),
          onSeeked: (B) => st(B.currentTarget),
          onPlay: (B) => {
            B.currentTarget.getRootNode().querySelectorAll("audio").forEach((rt) => {
              rt !== B.currentTarget && rt.pause();
            });
          }
        }
      )
    ] })
  ] });
}
function W1({
  session: r,
  onClose: x
}) {
  const A = [...r.recordings || []].sort(
    (o, _) => o.startedAt.localeCompare(_.startedAt)
  );
  return /* @__PURE__ */ c.jsxs(
    ra,
    {
      wide: !0,
      title: "录音片段",
      description: `${r.title} · 共 ${A.length} 段，每段最长 10 分钟`,
      onClose: x,
      children: [
        /* @__PURE__ */ c.jsx("div", { className: "recordings-list", children: A.length ? A.map((o, _) => /* @__PURE__ */ c.jsx(
          $1,
          {
            recording: o,
            index: _
          },
          o.id
        )) : /* @__PURE__ */ c.jsxs("div", { className: "recordings-empty", children: [
          /* @__PURE__ */ c.jsx(Hf, { size: 20 }),
          "暂时没有录音"
        ] }) }),
        /* @__PURE__ */ c.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: x, children: "关闭" }) })
      ]
    }
  );
}
function F1({
  project: r,
  onClose: x,
  onSave: A
}) {
  const [o, _] = R.useState(r?.name || ""), [D, U] = R.useState(r?.rootPath || ""), [O, N] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    ra,
    {
      title: r ? "编辑项目" : "添加项目",
      description: "Action 会在这个本地目录中交给 Codex。",
      onClose: x,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (b) => {
            b.preventDefault(), N(!0), A({ name: o, rootPath: D }).finally(() => N(!1));
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "项目名称" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                value: o,
                onChange: (b) => _(b.target.value),
                placeholder: "我的产品"
              }
            ),
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "本地项目目录" }),
            /* @__PURE__ */ c.jsxs("div", { className: "path-field", children: [
              /* @__PURE__ */ c.jsx(qf, { size: 16 }),
              /* @__PURE__ */ c.jsx(
                "input",
                {
                  value: D,
                  onChange: (b) => U(b.target.value),
                  placeholder: "C:\\\\Code\\\\my-project"
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: x, children: "取消" }),
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: "primary-button",
                  disabled: O || !o.trim() || !D.trim(),
                  children: [
                    O ? /* @__PURE__ */ c.jsx(ge, { size: 15 }) : /* @__PURE__ */ c.jsx(Wu, { size: 15 }),
                    " 保存项目"
                  ]
                }
              )
            ] })
          ]
        }
      )
    }
  );
}
function P1({
  session: r,
  onClose: x,
  onSave: A
}) {
  const [o, _] = R.useState(r.title), [D, U] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    ra,
    {
      title: "重命名会话",
      description: "日期时间名称也可以随时改掉。",
      onClose: x,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), U(!0), A(o).finally(() => U(!1));
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "会话名称" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                autoFocus: !0,
                value: o,
                onChange: (O) => _(O.target.value)
              }
            ),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: x, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: D || !o.trim(), children: [
                D ? /* @__PURE__ */ c.jsx(ge, { size: 15 }) : /* @__PURE__ */ c.jsx(Wu, { size: 15 }),
                " 保存"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function I1({
  action: r,
  onClose: x,
  onSave: A
}) {
  const [o, _] = R.useState(r?.title || ""), [D, U] = R.useState(r?.detail || ""), [O, N] = R.useState(
    r?.priority || "medium"
  ), [b, C] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    ra,
    {
      title: r ? "完善 Action" : "添加 Action",
      description: "把目标和验收结果说清楚，Codex 会做得更稳。",
      onClose: x,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (G) => {
            G.preventDefault(), C(!0), A({ title: o, detail: D, priority: O }).finally(
              () => C(!1)
            );
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "要做什么" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                value: o,
                autoFocus: !0,
                onChange: (G) => _(G.target.value),
                placeholder: "一句话描述结果"
              }
            ),
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "详细说明" }),
            /* @__PURE__ */ c.jsx(
              "textarea",
              {
                className: "text-field textarea",
                value: D,
                onChange: (G) => U(G.target.value),
                placeholder: "补充范围、约束和验收标准…"
              }
            ),
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "优先级" }),
            /* @__PURE__ */ c.jsx("div", { className: "priority-picker", children: ["high", "medium", "low"].map((G) => /* @__PURE__ */ c.jsxs(
              "button",
              {
                type: "button",
                className: O === G ? "active" : "",
                onClick: () => N(G),
                children: [
                  /* @__PURE__ */ c.jsx(h1, { size: 9, fill: "currentColor" }),
                  " ",
                  G === "high" ? "高" : G === "medium" ? "普通" : "低"
                ]
              },
              G
            )) }),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: x, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: b || !o.trim(), children: [
                b ? /* @__PURE__ */ c.jsx(ge, { size: 15 }) : /* @__PURE__ */ c.jsx(Wu, { size: 15 }),
                " 保存 Action"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function ty({
  action: r,
  state: x,
  onClose: A
}) {
  const [o, _] = R.useState("");
  r = x.actions.find((U) => U.id === r.id) || r;
  const D = x.runs.find((U) => U.id === r.runId);
  return /* @__PURE__ */ c.jsxs(
    ra,
    {
      wide: !0,
      title: r.status === "running" ? "Codex 正在工作" : r.status === "done" ? "Codex 已完成" : "Codex 执行结果",
      description: r.title,
      onClose: A,
      children: [
        /* @__PURE__ */ c.jsxs("div", { className: "run-meta", children: [
          /* @__PURE__ */ c.jsx(Qd, { status: r.status }),
          D && /* @__PURE__ */ c.jsxs("span", { children: [
            /* @__PURE__ */ c.jsx(y1, { size: 13 }),
            " ",
            new Date(D.startedAt).toLocaleString("zh-CN")
          ] })
        ] }),
        /* @__PURE__ */ c.jsx("pre", { className: "run-output", children: D?.output || (D?.status === "running" ? "正在启动 Codex…" : D?.error || "暂无输出") }),
        D?.error && /* @__PURE__ */ c.jsx("p", { role: "alert", children: D.error }),
        o && /* @__PURE__ */ c.jsx("p", { role: "alert", children: o }),
        /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
          D?.status === "running" && /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "ghost-button",
              onClick: () => {
                Qt.cancelRun(D.id).catch((U) => _(U.message));
              },
              children: "停止执行"
            }
          ),
          /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: A, children: "关闭" })
        ] })
      ]
    }
  );
}
function ly(r, x, A, o, _) {
  M1(x);
  const D = r.shadowRoot || r.attachShadow({ mode: "open" });
  D.replaceChildren();
  const U = document.createElement("link");
  U.rel = "stylesheet", U.href = new URL(
    /* @vite-ignore */
    "./sayso.css",
    import.meta.url
  ).href;
  const O = document.createElement("div");
  O.id = "root", D.append(U, O);
  const N = f1.createRoot(O);
  let b;
  r.toggleAttribute("data-embedded", !!_), N.render(/* @__PURE__ */ c.jsx(Q1, { onClose: A, deviceLabel: o, scope: _, registerClose: (G) => {
    b = G;
  } }));
  const C = () => N.unmount();
  return C.close = async () => b ? b() : !1, C;
}
export {
  ly as mount
};
