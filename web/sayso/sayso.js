var zf = { exports: {} }, $u = {};
var Td;
function I0() {
  if (Td) return $u;
  Td = 1;
  var o = /* @__PURE__ */ Symbol.for("react.transitional.element"), b = /* @__PURE__ */ Symbol.for("react.fragment");
  function j(r, _, D) {
    var U = null;
    if (D !== void 0 && (U = "" + D), _.key !== void 0 && (U = "" + _.key), "key" in _) {
      D = {};
      for (var O in _)
        O !== "key" && (D[O] = _[O]);
    } else D = _;
    return _ = D.ref, {
      $$typeof: o,
      type: r,
      key: U,
      ref: _ !== void 0 ? _ : null,
      props: D
    };
  }
  return $u.Fragment = b, $u.jsx = j, $u.jsxs = j, $u;
}
var Ed;
function t1() {
  return Ed || (Ed = 1, zf.exports = I0()), zf.exports;
}
var c = t1(), Mf = { exports: {} }, Wu = {}, Nf = { exports: {} }, Df = {};
var Ad;
function l1() {
  return Ad || (Ad = 1, (function(o) {
    function b(x, H) {
      var K = x.length;
      x.push(H);
      t: for (; 0 < K; ) {
        var dt = K - 1 >>> 1, d = x[dt];
        if (0 < _(d, H))
          x[dt] = H, x[K] = d, K = dt;
        else break t;
      }
    }
    function j(x) {
      return x.length === 0 ? null : x[0];
    }
    function r(x) {
      if (x.length === 0) return null;
      var H = x[0], K = x.pop();
      if (K !== H) {
        x[0] = K;
        t: for (var dt = 0, d = x.length, N = d >>> 1; dt < N; ) {
          var Y = 2 * (dt + 1) - 1, q = x[Y], w = Y + 1, ct = x[w];
          if (0 > _(q, K))
            w < d && 0 > _(ct, q) ? (x[dt] = ct, x[w] = K, dt = w) : (x[dt] = q, x[Y] = K, dt = Y);
          else if (w < d && 0 > _(ct, K))
            x[dt] = ct, x[w] = K, dt = w;
          else break t;
        }
      }
      return H;
    }
    function _(x, H) {
      var K = x.sortIndex - H.sortIndex;
      return K !== 0 ? K : x.id - H.id;
    }
    if (o.unstable_now = void 0, typeof performance == "object" && typeof performance.now == "function") {
      var D = performance;
      o.unstable_now = function() {
        return D.now();
      };
    } else {
      var U = Date, O = U.now();
      o.unstable_now = function() {
        return U.now() - O;
      };
    }
    var M = [], p = [], C = 1, G = null, Z = 3, tt = !1, B = !1, yt = !1, Nt = !1, Wt = typeof setTimeout == "function" ? setTimeout : null, pt = typeof clearTimeout == "function" ? clearTimeout : null, Et = typeof setImmediate < "u" ? setImmediate : null;
    function Bt(x) {
      for (var H = j(p); H !== null; ) {
        if (H.callback === null) r(p);
        else if (H.startTime <= x)
          r(p), H.sortIndex = H.expirationTime, b(M, H);
        else break;
        H = j(p);
      }
    }
    function P(x) {
      if (yt = !1, Bt(x), !B)
        if (j(M) !== null)
          B = !0, at || (at = !0, Qt());
        else {
          var H = j(p);
          H !== null && Zt(P, H.startTime - x);
        }
    }
    var at = !1, Ft = -1, V = 5, Ut = -1;
    function Bl() {
      return Nt ? !0 : !(o.unstable_now() - Ut < V);
    }
    function sl() {
      if (Nt = !1, at) {
        var x = o.unstable_now();
        Ut = x;
        var H = !0;
        try {
          t: {
            B = !1, yt && (yt = !1, pt(Ft), Ft = -1), tt = !0;
            var K = Z;
            try {
              l: {
                for (Bt(x), G = j(M); G !== null && !(G.expirationTime > x && Bl()); ) {
                  var dt = G.callback;
                  if (typeof dt == "function") {
                    G.callback = null, Z = G.priorityLevel;
                    var d = dt(
                      G.expirationTime <= x
                    );
                    if (x = o.unstable_now(), typeof d == "function") {
                      G.callback = d, Bt(x), H = !0;
                      break l;
                    }
                    G === j(M) && r(M), Bt(x);
                  } else r(M);
                  G = j(M);
                }
                if (G !== null) H = !0;
                else {
                  var N = j(p);
                  N !== null && Zt(
                    P,
                    N.startTime - x
                  ), H = !1;
                }
              }
              break t;
            } finally {
              G = null, Z = K, tt = !1;
            }
            H = void 0;
          }
        } finally {
          H ? Qt() : at = !1;
        }
      }
    }
    var Qt;
    if (typeof Et == "function")
      Qt = function() {
        Et(sl);
      };
    else if (typeof MessageChannel < "u") {
      var gl = new MessageChannel(), $l = gl.port2;
      gl.port1.onmessage = sl, Qt = function() {
        $l.postMessage(null);
      };
    } else
      Qt = function() {
        Wt(sl, 0);
      };
    function Zt(x, H) {
      Ft = Wt(function() {
        x(o.unstable_now());
      }, H);
    }
    o.unstable_IdlePriority = 5, o.unstable_ImmediatePriority = 1, o.unstable_LowPriority = 4, o.unstable_NormalPriority = 3, o.unstable_Profiling = null, o.unstable_UserBlockingPriority = 2, o.unstable_cancelCallback = function(x) {
      x.callback = null;
    }, o.unstable_forceFrameRate = function(x) {
      0 > x || 125 < x ? console.error(
        "forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"
      ) : V = 0 < x ? Math.floor(1e3 / x) : 5;
    }, o.unstable_getCurrentPriorityLevel = function() {
      return Z;
    }, o.unstable_next = function(x) {
      switch (Z) {
        case 1:
        case 2:
        case 3:
          var H = 3;
          break;
        default:
          H = Z;
      }
      var K = Z;
      Z = H;
      try {
        return x();
      } finally {
        Z = K;
      }
    }, o.unstable_requestPaint = function() {
      Nt = !0;
    }, o.unstable_runWithPriority = function(x, H) {
      switch (x) {
        case 1:
        case 2:
        case 3:
        case 4:
        case 5:
          break;
        default:
          x = 3;
      }
      var K = Z;
      Z = x;
      try {
        return H();
      } finally {
        Z = K;
      }
    }, o.unstable_scheduleCallback = function(x, H, K) {
      var dt = o.unstable_now();
      switch (typeof K == "object" && K !== null ? (K = K.delay, K = typeof K == "number" && 0 < K ? dt + K : dt) : K = dt, x) {
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
      return d = K + d, x = {
        id: C++,
        callback: H,
        priorityLevel: x,
        startTime: K,
        expirationTime: d,
        sortIndex: -1
      }, K > dt ? (x.sortIndex = K, b(p, x), j(M) === null && x === j(p) && (yt ? (pt(Ft), Ft = -1) : yt = !0, Zt(P, K - dt))) : (x.sortIndex = d, b(M, x), B || tt || (B = !0, at || (at = !0, Qt()))), x;
    }, o.unstable_shouldYield = Bl, o.unstable_wrapCallback = function(x) {
      var H = Z;
      return function() {
        var K = Z;
        Z = H;
        try {
          return x.apply(this, arguments);
        } finally {
          Z = K;
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
  var o = /* @__PURE__ */ Symbol.for("react.transitional.element"), b = /* @__PURE__ */ Symbol.for("react.portal"), j = /* @__PURE__ */ Symbol.for("react.fragment"), r = /* @__PURE__ */ Symbol.for("react.strict_mode"), _ = /* @__PURE__ */ Symbol.for("react.profiler"), D = /* @__PURE__ */ Symbol.for("react.consumer"), U = /* @__PURE__ */ Symbol.for("react.context"), O = /* @__PURE__ */ Symbol.for("react.forward_ref"), M = /* @__PURE__ */ Symbol.for("react.suspense"), p = /* @__PURE__ */ Symbol.for("react.memo"), C = /* @__PURE__ */ Symbol.for("react.lazy"), G = Symbol.iterator;
  function Z(d) {
    return d === null || typeof d != "object" ? null : (d = G && d[G] || d["@@iterator"], typeof d == "function" ? d : null);
  }
  var tt = {
    isMounted: function() {
      return !1;
    },
    enqueueForceUpdate: function() {
    },
    enqueueReplaceState: function() {
    },
    enqueueSetState: function() {
    }
  }, B = Object.assign, yt = {};
  function Nt(d, N, Y) {
    this.props = d, this.context = N, this.refs = yt, this.updater = Y || tt;
  }
  Nt.prototype.isReactComponent = {}, Nt.prototype.setState = function(d, N) {
    if (typeof d != "object" && typeof d != "function" && d != null)
      throw Error(
        "takes an object of state variables to update or a function which returns an object of state variables."
      );
    this.updater.enqueueSetState(this, d, N, "setState");
  }, Nt.prototype.forceUpdate = function(d) {
    this.updater.enqueueForceUpdate(this, d, "forceUpdate");
  };
  function Wt() {
  }
  Wt.prototype = Nt.prototype;
  function pt(d, N, Y) {
    this.props = d, this.context = N, this.refs = yt, this.updater = Y || tt;
  }
  var Et = pt.prototype = new Wt();
  Et.constructor = pt, B(Et, Nt.prototype), Et.isPureReactComponent = !0;
  var Bt = Array.isArray, P = { H: null, A: null, T: null, S: null, V: null }, at = Object.prototype.hasOwnProperty;
  function Ft(d, N, Y, q, w, ct) {
    return Y = ct.ref, {
      $$typeof: o,
      type: d,
      key: N,
      ref: Y !== void 0 ? Y : null,
      props: ct
    };
  }
  function V(d, N) {
    return Ft(
      d.type,
      N,
      void 0,
      void 0,
      void 0,
      d.props
    );
  }
  function Ut(d) {
    return typeof d == "object" && d !== null && d.$$typeof === o;
  }
  function Bl(d) {
    var N = { "=": "=0", ":": "=2" };
    return "$" + d.replace(/[=:]/g, function(Y) {
      return N[Y];
    });
  }
  var sl = /\/+/g;
  function Qt(d, N) {
    return typeof d == "object" && d !== null && d.key != null ? Bl("" + d.key) : N.toString(36);
  }
  function gl() {
  }
  function $l(d) {
    switch (d.status) {
      case "fulfilled":
        return d.value;
      case "rejected":
        throw d.reason;
      default:
        switch (typeof d.status == "string" ? d.then(gl, gl) : (d.status = "pending", d.then(
          function(N) {
            d.status === "pending" && (d.status = "fulfilled", d.value = N);
          },
          function(N) {
            d.status === "pending" && (d.status = "rejected", d.reason = N);
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
  function Zt(d, N, Y, q, w) {
    var ct = typeof d;
    (ct === "undefined" || ct === "boolean") && (d = null);
    var $ = !1;
    if (d === null) $ = !0;
    else
      switch (ct) {
        case "bigint":
        case "string":
        case "number":
          $ = !0;
          break;
        case "object":
          switch (d.$$typeof) {
            case o:
            case b:
              $ = !0;
              break;
            case C:
              return $ = d._init, Zt(
                $(d._payload),
                N,
                Y,
                q,
                w
              );
          }
      }
    if ($)
      return w = w(d), $ = q === "" ? "." + Qt(d, 0) : q, Bt(w) ? (Y = "", $ != null && (Y = $.replace(sl, "$&/") + "/"), Zt(w, N, Y, "", function(Yl) {
        return Yl;
      })) : w != null && (Ut(w) && (w = V(
        w,
        Y + (w.key == null || d && d.key === w.key ? "" : ("" + w.key).replace(
          sl,
          "$&/"
        ) + "/") + $
      )), N.push(w)), 1;
    $ = 0;
    var el = q === "" ? "." : q + ":";
    if (Bt(d))
      for (var xt = 0; xt < d.length; xt++)
        q = d[xt], ct = el + Qt(q, xt), $ += Zt(
          q,
          N,
          Y,
          ct,
          w
        );
    else if (xt = Z(d), typeof xt == "function")
      for (d = xt.call(d), xt = 0; !(q = d.next()).done; )
        q = q.value, ct = el + Qt(q, xt++), $ += Zt(
          q,
          N,
          Y,
          ct,
          w
        );
    else if (ct === "object") {
      if (typeof d.then == "function")
        return Zt(
          $l(d),
          N,
          Y,
          q,
          w
        );
      throw N = String(d), Error(
        "Objects are not valid as a React child (found: " + (N === "[object Object]" ? "object with keys {" + Object.keys(d).join(", ") + "}" : N) + "). If you meant to render a collection of children, use an array instead."
      );
    }
    return $;
  }
  function x(d, N, Y) {
    if (d == null) return d;
    var q = [], w = 0;
    return Zt(d, q, "", "", function(ct) {
      return N.call(Y, ct, w++);
    }), q;
  }
  function H(d) {
    if (d._status === -1) {
      var N = d._result;
      N = N(), N.then(
        function(Y) {
          (d._status === 0 || d._status === -1) && (d._status = 1, d._result = Y);
        },
        function(Y) {
          (d._status === 0 || d._status === -1) && (d._status = 2, d._result = Y);
        }
      ), d._status === -1 && (d._status = 0, d._result = N);
    }
    if (d._status === 1) return d._result.default;
    throw d._result;
  }
  var K = typeof reportError == "function" ? reportError : function(d) {
    if (typeof window == "object" && typeof window.ErrorEvent == "function") {
      var N = new window.ErrorEvent("error", {
        bubbles: !0,
        cancelable: !0,
        message: typeof d == "object" && d !== null && typeof d.message == "string" ? String(d.message) : String(d),
        error: d
      });
      if (!window.dispatchEvent(N)) return;
    } else if (typeof process == "object" && typeof process.emit == "function") {
      process.emit("uncaughtException", d);
      return;
    }
    console.error(d);
  };
  function dt() {
  }
  return I.Children = {
    map: x,
    forEach: function(d, N, Y) {
      x(
        d,
        function() {
          N.apply(this, arguments);
        },
        Y
      );
    },
    count: function(d) {
      var N = 0;
      return x(d, function() {
        N++;
      }), N;
    },
    toArray: function(d) {
      return x(d, function(N) {
        return N;
      }) || [];
    },
    only: function(d) {
      if (!Ut(d))
        throw Error(
          "React.Children.only expected to receive a single React element child."
        );
      return d;
    }
  }, I.Component = Nt, I.Fragment = j, I.Profiler = _, I.PureComponent = pt, I.StrictMode = r, I.Suspense = M, I.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = P, I.__COMPILER_RUNTIME = {
    __proto__: null,
    c: function(d) {
      return P.H.useMemoCache(d);
    }
  }, I.cache = function(d) {
    return function() {
      return d.apply(null, arguments);
    };
  }, I.cloneElement = function(d, N, Y) {
    if (d == null)
      throw Error(
        "The argument must be a React element, but you passed " + d + "."
      );
    var q = B({}, d.props), w = d.key, ct = void 0;
    if (N != null)
      for ($ in N.ref !== void 0 && (ct = void 0), N.key !== void 0 && (w = "" + N.key), N)
        !at.call(N, $) || $ === "key" || $ === "__self" || $ === "__source" || $ === "ref" && N.ref === void 0 || (q[$] = N[$]);
    var $ = arguments.length - 2;
    if ($ === 1) q.children = Y;
    else if (1 < $) {
      for (var el = Array($), xt = 0; xt < $; xt++)
        el[xt] = arguments[xt + 2];
      q.children = el;
    }
    return Ft(d.type, w, void 0, void 0, ct, q);
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
  }, I.createElement = function(d, N, Y) {
    var q, w = {}, ct = null;
    if (N != null)
      for (q in N.key !== void 0 && (ct = "" + N.key), N)
        at.call(N, q) && q !== "key" && q !== "__self" && q !== "__source" && (w[q] = N[q]);
    var $ = arguments.length - 2;
    if ($ === 1) w.children = Y;
    else if (1 < $) {
      for (var el = Array($), xt = 0; xt < $; xt++)
        el[xt] = arguments[xt + 2];
      w.children = el;
    }
    if (d && d.defaultProps)
      for (q in $ = d.defaultProps, $)
        w[q] === void 0 && (w[q] = $[q]);
    return Ft(d, ct, void 0, void 0, null, w);
  }, I.createRef = function() {
    return { current: null };
  }, I.forwardRef = function(d) {
    return { $$typeof: O, render: d };
  }, I.isValidElement = Ut, I.lazy = function(d) {
    return {
      $$typeof: C,
      _payload: { _status: -1, _result: d },
      _init: H
    };
  }, I.memo = function(d, N) {
    return {
      $$typeof: p,
      type: d,
      compare: N === void 0 ? null : N
    };
  }, I.startTransition = function(d) {
    var N = P.T, Y = {};
    P.T = Y;
    try {
      var q = d(), w = P.S;
      w !== null && w(Y, q), typeof q == "object" && q !== null && typeof q.then == "function" && q.then(dt, K);
    } catch (ct) {
      K(ct);
    } finally {
      P.T = N;
    }
  }, I.unstable_useCacheRefresh = function() {
    return P.H.useCacheRefresh();
  }, I.use = function(d) {
    return P.H.use(d);
  }, I.useActionState = function(d, N, Y) {
    return P.H.useActionState(d, N, Y);
  }, I.useCallback = function(d, N) {
    return P.H.useCallback(d, N);
  }, I.useContext = function(d) {
    return P.H.useContext(d);
  }, I.useDebugValue = function() {
  }, I.useDeferredValue = function(d, N) {
    return P.H.useDeferredValue(d, N);
  }, I.useEffect = function(d, N, Y) {
    var q = P.H;
    if (typeof Y == "function")
      throw Error(
        "useEffect CRUD overload is not enabled in this build of React."
      );
    return q.useEffect(d, N);
  }, I.useId = function() {
    return P.H.useId();
  }, I.useImperativeHandle = function(d, N, Y) {
    return P.H.useImperativeHandle(d, N, Y);
  }, I.useInsertionEffect = function(d, N) {
    return P.H.useInsertionEffect(d, N);
  }, I.useLayoutEffect = function(d, N) {
    return P.H.useLayoutEffect(d, N);
  }, I.useMemo = function(d, N) {
    return P.H.useMemo(d, N);
  }, I.useOptimistic = function(d, N) {
    return P.H.useOptimistic(d, N);
  }, I.useReducer = function(d, N, Y) {
    return P.H.useReducer(d, N, Y);
  }, I.useRef = function(d) {
    return P.H.useRef(d);
  }, I.useState = function(d) {
    return P.H.useState(d);
  }, I.useSyncExternalStore = function(d, N, Y) {
    return P.H.useSyncExternalStore(
      d,
      N,
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
var Rf = { exports: {} }, nl = {};
var Dd;
function u1() {
  if (Dd) return nl;
  Dd = 1;
  var o = Uf();
  function b(M) {
    var p = "https://react.dev/errors/" + M;
    if (1 < arguments.length) {
      p += "?args[]=" + encodeURIComponent(arguments[1]);
      for (var C = 2; C < arguments.length; C++)
        p += "&args[]=" + encodeURIComponent(arguments[C]);
    }
    return "Minified React error #" + M + "; visit " + p + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
  }
  function j() {
  }
  var r = {
    d: {
      f: j,
      r: function() {
        throw Error(b(522));
      },
      D: j,
      C: j,
      L: j,
      m: j,
      X: j,
      S: j,
      M: j
    },
    p: 0,
    findDOMNode: null
  }, _ = /* @__PURE__ */ Symbol.for("react.portal");
  function D(M, p, C) {
    var G = 3 < arguments.length && arguments[3] !== void 0 ? arguments[3] : null;
    return {
      $$typeof: _,
      key: G == null ? null : "" + G,
      children: M,
      containerInfo: p,
      implementation: C
    };
  }
  var U = o.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;
  function O(M, p) {
    if (M === "font") return "";
    if (typeof p == "string")
      return p === "use-credentials" ? p : "";
  }
  return nl.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = r, nl.createPortal = function(M, p) {
    var C = 2 < arguments.length && arguments[2] !== void 0 ? arguments[2] : null;
    if (!p || p.nodeType !== 1 && p.nodeType !== 9 && p.nodeType !== 11)
      throw Error(b(299));
    return D(M, p, null, C);
  }, nl.flushSync = function(M) {
    var p = U.T, C = r.p;
    try {
      if (U.T = null, r.p = 2, M) return M();
    } finally {
      U.T = p, r.p = C, r.d.f();
    }
  }, nl.preconnect = function(M, p) {
    typeof M == "string" && (p ? (p = p.crossOrigin, p = typeof p == "string" ? p === "use-credentials" ? p : "" : void 0) : p = null, r.d.C(M, p));
  }, nl.prefetchDNS = function(M) {
    typeof M == "string" && r.d.D(M);
  }, nl.preinit = function(M, p) {
    if (typeof M == "string" && p && typeof p.as == "string") {
      var C = p.as, G = O(C, p.crossOrigin), Z = typeof p.integrity == "string" ? p.integrity : void 0, tt = typeof p.fetchPriority == "string" ? p.fetchPriority : void 0;
      C === "style" ? r.d.S(
        M,
        typeof p.precedence == "string" ? p.precedence : void 0,
        {
          crossOrigin: G,
          integrity: Z,
          fetchPriority: tt
        }
      ) : C === "script" && r.d.X(M, {
        crossOrigin: G,
        integrity: Z,
        fetchPriority: tt,
        nonce: typeof p.nonce == "string" ? p.nonce : void 0
      });
    }
  }, nl.preinitModule = function(M, p) {
    if (typeof M == "string")
      if (typeof p == "object" && p !== null) {
        if (p.as == null || p.as === "script") {
          var C = O(
            p.as,
            p.crossOrigin
          );
          r.d.M(M, {
            crossOrigin: C,
            integrity: typeof p.integrity == "string" ? p.integrity : void 0,
            nonce: typeof p.nonce == "string" ? p.nonce : void 0
          });
        }
      } else p == null && r.d.M(M);
  }, nl.preload = function(M, p) {
    if (typeof M == "string" && typeof p == "object" && p !== null && typeof p.as == "string") {
      var C = p.as, G = O(C, p.crossOrigin);
      r.d.L(M, C, {
        crossOrigin: G,
        integrity: typeof p.integrity == "string" ? p.integrity : void 0,
        nonce: typeof p.nonce == "string" ? p.nonce : void 0,
        type: typeof p.type == "string" ? p.type : void 0,
        fetchPriority: typeof p.fetchPriority == "string" ? p.fetchPriority : void 0,
        referrerPolicy: typeof p.referrerPolicy == "string" ? p.referrerPolicy : void 0,
        imageSrcSet: typeof p.imageSrcSet == "string" ? p.imageSrcSet : void 0,
        imageSizes: typeof p.imageSizes == "string" ? p.imageSizes : void 0,
        media: typeof p.media == "string" ? p.media : void 0
      });
    }
  }, nl.preloadModule = function(M, p) {
    if (typeof M == "string")
      if (p) {
        var C = O(p.as, p.crossOrigin);
        r.d.m(M, {
          as: typeof p.as == "string" && p.as !== "script" ? p.as : void 0,
          crossOrigin: C,
          integrity: typeof p.integrity == "string" ? p.integrity : void 0
        });
      } else r.d.m(M);
  }, nl.requestFormReset = function(M) {
    r.d.r(M);
  }, nl.unstable_batchedUpdates = function(M, p) {
    return M(p);
  }, nl.useFormState = function(M, p, C) {
    return U.H.useFormState(M, p, C);
  }, nl.useFormStatus = function() {
    return U.H.useHostTransitionStatus();
  }, nl.version = "19.1.1", nl;
}
var Od;
function n1() {
  if (Od) return Rf.exports;
  Od = 1;
  function o() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o);
      } catch (b) {
        console.error(b);
      }
  }
  return o(), Rf.exports = u1(), Rf.exports;
}
var Rd;
function i1() {
  if (Rd) return Wu;
  Rd = 1;
  var o = e1(), b = Uf(), j = n1();
  function r(t) {
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
      throw Error(r(188));
  }
  function M(t) {
    var l = t.alternate;
    if (!l) {
      if (l = D(t), l === null) throw Error(r(188));
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
        throw Error(r(188));
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
          if (!i) throw Error(r(189));
        }
      }
      if (e.alternate !== a) throw Error(r(190));
    }
    if (e.tag !== 3) throw Error(r(188));
    return e.stateNode.current === e ? t : l;
  }
  function p(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t;
    for (t = t.child; t !== null; ) {
      if (l = p(t), l !== null) return l;
      t = t.sibling;
    }
    return null;
  }
  var C = Object.assign, G = /* @__PURE__ */ Symbol.for("react.element"), Z = /* @__PURE__ */ Symbol.for("react.transitional.element"), tt = /* @__PURE__ */ Symbol.for("react.portal"), B = /* @__PURE__ */ Symbol.for("react.fragment"), yt = /* @__PURE__ */ Symbol.for("react.strict_mode"), Nt = /* @__PURE__ */ Symbol.for("react.profiler"), Wt = /* @__PURE__ */ Symbol.for("react.provider"), pt = /* @__PURE__ */ Symbol.for("react.consumer"), Et = /* @__PURE__ */ Symbol.for("react.context"), Bt = /* @__PURE__ */ Symbol.for("react.forward_ref"), P = /* @__PURE__ */ Symbol.for("react.suspense"), at = /* @__PURE__ */ Symbol.for("react.suspense_list"), Ft = /* @__PURE__ */ Symbol.for("react.memo"), V = /* @__PURE__ */ Symbol.for("react.lazy"), Ut = /* @__PURE__ */ Symbol.for("react.activity"), Bl = /* @__PURE__ */ Symbol.for("react.memo_cache_sentinel"), sl = Symbol.iterator;
  function Qt(t) {
    return t === null || typeof t != "object" ? null : (t = sl && t[sl] || t["@@iterator"], typeof t == "function" ? t : null);
  }
  var gl = /* @__PURE__ */ Symbol.for("react.client.reference");
  function $l(t) {
    if (t == null) return null;
    if (typeof t == "function")
      return t.$$typeof === gl ? null : t.displayName || t.name || null;
    if (typeof t == "string") return t;
    switch (t) {
      case B:
        return "Fragment";
      case Nt:
        return "Profiler";
      case yt:
        return "StrictMode";
      case P:
        return "Suspense";
      case at:
        return "SuspenseList";
      case Ut:
        return "Activity";
    }
    if (typeof t == "object")
      switch (t.$$typeof) {
        case tt:
          return "Portal";
        case Et:
          return (t.displayName || "Context") + ".Provider";
        case pt:
          return (t._context.displayName || "Context") + ".Consumer";
        case Bt:
          var l = t.render;
          return t = t.displayName, t || (t = l.displayName || l.name || "", t = t !== "" ? "ForwardRef(" + t + ")" : "ForwardRef"), t;
        case Ft:
          return l = t.displayName || null, l !== null ? l : $l(t.type) || "Memo";
        case V:
          l = t._payload, t = t._init;
          try {
            return $l(t(l));
          } catch {
          }
      }
    return null;
  }
  var Zt = Array.isArray, x = b.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, H = j.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, K = {
    pending: !1,
    data: null,
    method: null,
    action: null
  }, dt = [], d = -1;
  function N(t) {
    return { current: t };
  }
  function Y(t) {
    0 > d || (t.current = dt[d], dt[d] = null, d--);
  }
  function q(t, l) {
    d++, dt[d] = t.current, t.current = l;
  }
  var w = N(null), ct = N(null), $ = N(null), el = N(null);
  function xt(t, l) {
    switch (q($, l), q(ct, t), q(w, null), l.nodeType) {
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
    Y(w), q(w, t);
  }
  function Yl() {
    Y(w), Y(ct), Y($);
  }
  function Pa(t) {
    t.memoizedState !== null && q(el, t);
    var l = w.current, e = Po(l, t.type);
    l !== e && (q(ct, t), q(w, e));
  }
  function ra(t) {
    ct.current === t && (Y(w), Y(ct)), el.current === t && (Y(el), Vu._currentValue = K);
  }
  var Ia = Object.prototype.hasOwnProperty, Xe = o.unstable_scheduleCallback, Sl = o.unstable_cancelCallback, Pu = o.unstable_shouldYield, Qe = o.unstable_requestPaint, bl = o.unstable_now, vi = o.unstable_getCurrentPriorityLevel, tu = o.unstable_ImmediatePriority, rl = o.unstable_UserBlockingPriority, Wl = o.unstable_NormalPriority, ye = o.unstable_LowPriority, Zl = o.unstable_IdlePriority, pl = o.log, Ze = o.unstable_setDisableYieldValue, ol = null, ft = null;
  function Dt(t) {
    if (typeof pl == "function" && Ze(t), ft && typeof ft.setStrictMode == "function")
      try {
        ft.setStrictMode(ol, t);
      } catch {
      }
  }
  var Pt = Math.clz32 ? Math.clz32 : Iu, ve = Math.log, Ot = Math.LN2;
  function Iu(t) {
    return t >>>= 0, t === 0 ? 32 : 31 - (ve(t) / Ot | 0) | 0;
  }
  var oa = 256, da = 4194304;
  function Ll(t) {
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
  function Le(t, l, e) {
    var a = t.pendingLanes;
    if (a === 0) return 0;
    var u = 0, n = t.suspendedLanes, i = t.pingedLanes;
    t = t.warmLanes;
    var f = a & 134217727;
    return f !== 0 ? (a = f & ~n, a !== 0 ? u = Ll(a) : (i &= f, i !== 0 ? u = Ll(i) : e || (e = f & ~t, e !== 0 && (u = Ll(e))))) : (f = a & ~n, f !== 0 ? u = Ll(f) : i !== 0 ? u = Ll(i) : e || (e = a & ~t, e !== 0 && (u = Ll(e)))), u === 0 ? 0 : l !== 0 && l !== u && (l & n) === 0 && (n = u & -u, e = l & -l, n >= e || n === 32 && (e & 4194048) !== 0) ? l : u;
  }
  function Ve(t, l) {
    return (t.pendingLanes & ~(t.suspendedLanes & ~t.pingedLanes) & l) === 0;
  }
  function mi(t, l) {
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
  function we() {
    var t = oa;
    return oa <<= 1, (oa & 4194048) === 0 && (oa = 256), t;
  }
  function me() {
    var t = da;
    return da <<= 1, (da & 62914560) === 0 && (da = 4194304), t;
  }
  function E(t) {
    for (var l = [], e = 0; 31 > e; e++) l.push(t);
    return l;
  }
  function X(t, l) {
    t.pendingLanes |= l, l !== 268435456 && (t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0);
  }
  function F(t, l, e, a, u, n) {
    var i = t.pendingLanes;
    t.pendingLanes = e, t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0, t.expiredLanes &= e, t.entangledLanes &= e, t.errorRecoveryDisabledLanes &= e, t.shellSuspendCounter = 0;
    var f = t.entanglements, s = t.expirationTimes, m = t.hiddenUpdates;
    for (e = i & ~e; 0 < e; ) {
      var T = 31 - Pt(e), z = 1 << T;
      f[T] = 0, s[T] = -1;
      var g = m[T];
      if (g !== null)
        for (m[T] = null, T = 0; T < g.length; T++) {
          var S = g[T];
          S !== null && (S.lane &= -536870913);
        }
      e &= ~z;
    }
    a !== 0 && it(t, a, 0), n !== 0 && u === 0 && t.tag !== 0 && (t.suspendedLanes |= n & ~(i & ~l));
  }
  function it(t, l, e) {
    t.pendingLanes |= l, t.suspendedLanes &= ~l;
    var a = 31 - Pt(l);
    t.entangledLanes |= l, t.entanglements[a] = t.entanglements[a] | 1073741824 | e & 4194090;
  }
  function rt(t, l) {
    var e = t.entangledLanes |= l;
    for (t = t.entanglements; e; ) {
      var a = 31 - Pt(e), u = 1 << a;
      u & l | t[a] & l && (t[a] |= l), e &= ~u;
    }
  }
  function Lt(t) {
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
  function ge(t) {
    return t &= -t, 2 < t ? 8 < t ? (t & 134217727) !== 0 ? 32 : 268435456 : 8 : 2;
  }
  function dl() {
    var t = H.p;
    return t !== 0 ? t : (t = window.event, t === void 0 ? 32 : gd(t.type));
  }
  function Fl(t, l) {
    var e = H.p;
    try {
      return H.p = t, l();
    } finally {
      H.p = e;
    }
  }
  var il = Math.random().toString(36).slice(2), Yt = "__reactFiber$" + il, al = "__reactProps$" + il, ha = "__reactContainer$" + il, gi = "__reactEvents$" + il, Ld = "__reactListeners$" + il, Vd = "__reactHandles$" + il, Yf = "__reactResources$" + il, lu = "__reactMarker$" + il;
  function Si(t) {
    delete t[Yt], delete t[al], delete t[gi], delete t[Ld], delete t[Vd];
  }
  function ya(t) {
    var l = t[Yt];
    if (l) return l;
    for (var e = t.parentNode; e; ) {
      if (l = e[ha] || e[Yt]) {
        if (e = l.alternate, l.child !== null || e !== null && e.child !== null)
          for (t = ed(t); t !== null; ) {
            if (e = t[Yt]) return e;
            t = ed(t);
          }
        return l;
      }
      t = e, e = t.parentNode;
    }
    return null;
  }
  function va(t) {
    if (t = t[Yt] || t[ha]) {
      var l = t.tag;
      if (l === 5 || l === 6 || l === 13 || l === 26 || l === 27 || l === 3)
        return t;
    }
    return null;
  }
  function eu(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t.stateNode;
    throw Error(r(33));
  }
  function ma(t) {
    var l = t[Yf];
    return l || (l = t[Yf] = { hoistableStyles: /* @__PURE__ */ new Map(), hoistableScripts: /* @__PURE__ */ new Map() }), l;
  }
  function Kt(t) {
    t[lu] = !0;
  }
  var Gf = /* @__PURE__ */ new Set(), Xf = {};
  function Ke(t, l) {
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
    return Ia.call(Zf, t) ? !0 : Ia.call(Qf, t) ? !1 : wd.test(t) ? Zf[t] = !0 : (Qf[t] = !0, !1);
  }
  function tn(t, l, e) {
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
  function ln(t, l, e) {
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
  function Pl(t, l, e, a) {
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
  var pi = !1;
  function xi(t, l) {
    if (!t || pi) return "";
    pi = !0;
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
                  var T = `
` + s[a].replace(" at new ", " at ");
                  return t.displayName && T.includes("<anonymous>") && (T = T.replace("<anonymous>", t.displayName)), T;
                }
              while (1 <= a && 0 <= u);
            break;
          }
      }
    } finally {
      pi = !1, Error.prepareStackTrace = e;
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
        return xi(t.type, !1);
      case 11:
        return xi(t.type.render, !1);
      case 1:
        return xi(t.type, !0);
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
  function en(t) {
    t._valueTracker || (t._valueTracker = kd(t));
  }
  function Kf(t) {
    if (!t) return !1;
    var l = t._valueTracker;
    if (!l) return !0;
    var e = l.getValue(), a = "";
    return t && (a = wf(t) ? t.checked ? "true" : "false" : t.value), t = a, t !== e ? (l.setValue(t), !0) : !1;
  }
  function an(t) {
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
    l === "number" && an(t.ownerDocument) === t || t.defaultValue === "" + e || (t.defaultValue = "" + e);
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
        if (e != null) throw Error(r(92));
        if (Zt(a)) {
          if (1 < a.length) throw Error(r(93));
          a = a[0];
        }
        e = a;
      }
      e == null && (e = ""), l = e;
    }
    e = Nl(l), t.defaultValue = e, a = t.textContent, a === e && a !== "" && a !== null && (t.value = a);
  }
  function pa(t, l) {
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
      throw Error(r(62));
    if (t = t.style, e != null) {
      for (var a in e)
        !e.hasOwnProperty(a) || l != null && l.hasOwnProperty(a) || (a.indexOf("--") === 0 ? t.setProperty(a, "") : a === "float" ? t.cssFloat = "" : t[a] = "");
      for (var u in l)
        a = l[u], l.hasOwnProperty(u) && e[u] !== a && Wf(t, u, a);
    } else
      for (var n in l)
        l.hasOwnProperty(n) && Wf(t, n, l[n]);
  }
  function Ei(t) {
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
  function un(t) {
    return Pd.test("" + t) ? "javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')" : t;
  }
  var Ai = null;
  function zi(t) {
    return t = t.target || t.srcElement || window, t.correspondingUseElement && (t = t.correspondingUseElement), t.nodeType === 3 ? t.parentNode : t;
  }
  var xa = null, ja = null;
  function Pf(t) {
    var l = va(t);
    if (l && (t = l.stateNode)) {
      var e = t[al] || null;
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
                var u = a[al] || null;
                if (!u) throw Error(r(90));
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
      if (Mi = !1, (xa !== null || ja !== null) && (Vn(), xa && (l = xa, t = ja, ja = xa = null, Pf(l), t)))
        for (l = 0; l < t.length; l++) Pf(t[l]);
    }
  }
  function au(t, l) {
    var e = t.stateNode;
    if (e === null) return null;
    var a = e[al] || null;
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
        r(231, l, typeof e)
      );
    return e;
  }
  var Il = !(typeof window > "u" || typeof window.document > "u" || typeof window.document.createElement > "u"), Ni = !1;
  if (Il)
    try {
      var uu = {};
      Object.defineProperty(uu, "passive", {
        get: function() {
          Ni = !0;
        }
      }), window.addEventListener("test", uu, uu), window.removeEventListener("test", uu, uu);
    } catch {
      Ni = !1;
    }
  var Se = null, Di = null, nn = null;
  function ts() {
    if (nn) return nn;
    var t, l = Di, e = l.length, a, u = "value" in Se ? Se.value : Se.textContent, n = u.length;
    for (t = 0; t < e && l[t] === u[t]; t++) ;
    var i = e - t;
    for (a = 1; a <= i && l[e - a] === u[n - a]; a++) ;
    return nn = u.slice(t, 1 < a ? 1 - a : void 0);
  }
  function cn(t) {
    var l = t.keyCode;
    return "charCode" in t ? (t = t.charCode, t === 0 && l === 13 && (t = 13)) : t = l, t === 10 && (t = 13), 32 <= t || t === 13 ? t : 0;
  }
  function fn() {
    return !0;
  }
  function ls() {
    return !1;
  }
  function hl(t) {
    function l(e, a, u, n, i) {
      this._reactName = e, this._targetInst = u, this.type = a, this.nativeEvent = n, this.target = i, this.currentTarget = null;
      for (var f in t)
        t.hasOwnProperty(f) && (e = t[f], this[f] = e ? e(n) : n[f]);
      return this.isDefaultPrevented = (n.defaultPrevented != null ? n.defaultPrevented : n.returnValue === !1) ? fn : ls, this.isPropagationStopped = ls, this;
    }
    return C(l.prototype, {
      preventDefault: function() {
        this.defaultPrevented = !0;
        var e = this.nativeEvent;
        e && (e.preventDefault ? e.preventDefault() : typeof e.returnValue != "unknown" && (e.returnValue = !1), this.isDefaultPrevented = fn);
      },
      stopPropagation: function() {
        var e = this.nativeEvent;
        e && (e.stopPropagation ? e.stopPropagation() : typeof e.cancelBubble != "unknown" && (e.cancelBubble = !0), this.isPropagationStopped = fn);
      },
      persist: function() {
      },
      isPersistent: fn
    }), l;
  }
  var Je = {
    eventPhase: 0,
    bubbles: 0,
    cancelable: 0,
    timeStamp: function(t) {
      return t.timeStamp || Date.now();
    },
    defaultPrevented: 0,
    isTrusted: 0
  }, sn = hl(Je), nu = C({}, Je, { view: 0, detail: 0 }), Id = hl(nu), Oi, Ri, iu, rn = C({}, nu, {
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
      return "movementX" in t ? t.movementX : (t !== iu && (iu && t.type === "mousemove" ? (Oi = t.screenX - iu.screenX, Ri = t.screenY - iu.screenY) : Ri = Oi = 0, iu = t), Oi);
    },
    movementY: function(t) {
      return "movementY" in t ? t.movementY : Ri;
    }
  }), es = hl(rn), th = C({}, rn, { dataTransfer: 0 }), lh = hl(th), eh = C({}, nu, { relatedTarget: 0 }), _i = hl(eh), ah = C({}, Je, {
    animationName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), uh = hl(ah), nh = C({}, Je, {
    clipboardData: function(t) {
      return "clipboardData" in t ? t.clipboardData : window.clipboardData;
    }
  }), ih = hl(nh), ch = C({}, Je, { data: 0 }), as = hl(ch), fh = {
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
  var dh = C({}, nu, {
    key: function(t) {
      if (t.key) {
        var l = fh[t.key] || t.key;
        if (l !== "Unidentified") return l;
      }
      return t.type === "keypress" ? (t = cn(t), t === 13 ? "Enter" : String.fromCharCode(t)) : t.type === "keydown" || t.type === "keyup" ? sh[t.keyCode] || "Unidentified" : "";
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
      return t.type === "keypress" ? cn(t) : 0;
    },
    keyCode: function(t) {
      return t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    },
    which: function(t) {
      return t.type === "keypress" ? cn(t) : t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    }
  }), hh = hl(dh), yh = C({}, rn, {
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
  }), us = hl(yh), vh = C({}, nu, {
    touches: 0,
    targetTouches: 0,
    changedTouches: 0,
    altKey: 0,
    metaKey: 0,
    ctrlKey: 0,
    shiftKey: 0,
    getModifierState: Ui
  }), mh = hl(vh), gh = C({}, Je, {
    propertyName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), Sh = hl(gh), bh = C({}, rn, {
    deltaX: function(t) {
      return "deltaX" in t ? t.deltaX : "wheelDeltaX" in t ? -t.wheelDeltaX : 0;
    },
    deltaY: function(t) {
      return "deltaY" in t ? t.deltaY : "wheelDeltaY" in t ? -t.wheelDeltaY : "wheelDelta" in t ? -t.wheelDelta : 0;
    },
    deltaZ: 0,
    deltaMode: 0
  }), ph = hl(bh), xh = C({}, Je, {
    newState: 0,
    oldState: 0
  }), jh = hl(xh), Th = [9, 13, 27, 32], Ci = Il && "CompositionEvent" in window, cu = null;
  Il && "documentMode" in document && (cu = document.documentMode);
  var Eh = Il && "TextEvent" in window && !cu, ns = Il && (!Ci || cu && 8 < cu && 11 >= cu), is = " ", cs = !1;
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
  function Ah(t, l) {
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
      return t === "compositionend" || !Ci && fs(t, l) ? (t = ts(), nn = Di = Se = null, Ta = !1, t) : null;
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
    xa ? ja ? ja.push(a) : ja = [a] : xa = a, l = Wn(l, "onChange"), 0 < l.length && (e = new sn(
      "onChange",
      "change",
      null,
      e,
      a
    ), t.push({ event: e, listeners: l }));
  }
  var fu = null, su = null;
  function Nh(t) {
    Ko(t, 0);
  }
  function on(t) {
    var l = eu(t);
    if (Kf(l)) return t;
  }
  function ds(t, l) {
    if (t === "change") return l;
  }
  var hs = !1;
  if (Il) {
    var Hi;
    if (Il) {
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
    fu && (fu.detachEvent("onpropertychange", ms), su = fu = null);
  }
  function ms(t) {
    if (t.propertyName === "value" && on(su)) {
      var l = [];
      os(
        l,
        su,
        t,
        zi(t)
      ), If(Nh, l);
    }
  }
  function Dh(t, l, e) {
    t === "focusin" ? (vs(), fu = l, su = e, fu.attachEvent("onpropertychange", ms)) : t === "focusout" && vs();
  }
  function Oh(t) {
    if (t === "selectionchange" || t === "keyup" || t === "keydown")
      return on(su);
  }
  function Rh(t, l) {
    if (t === "click") return on(l);
  }
  function _h(t, l) {
    if (t === "input" || t === "change")
      return on(l);
  }
  function Uh(t, l) {
    return t === l && (t !== 0 || 1 / t === 1 / l) || t !== t && l !== l;
  }
  var xl = typeof Object.is == "function" ? Object.is : Uh;
  function ru(t, l) {
    if (xl(t, l)) return !0;
    if (typeof t != "object" || t === null || typeof l != "object" || l === null)
      return !1;
    var e = Object.keys(t), a = Object.keys(l);
    if (e.length !== a.length) return !1;
    for (a = 0; a < e.length; a++) {
      var u = e[a];
      if (!Ia.call(l, u) || !xl(t[u], l[u]))
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
  function ps(t) {
    t = t != null && t.ownerDocument != null && t.ownerDocument.defaultView != null ? t.ownerDocument.defaultView : window;
    for (var l = an(t.document); l instanceof t.HTMLIFrameElement; ) {
      try {
        var e = typeof l.contentWindow.location.href == "string";
      } catch {
        e = !1;
      }
      if (e) t = l.contentWindow;
      else break;
      l = an(t.document);
    }
    return l;
  }
  function Bi(t) {
    var l = t && t.nodeName && t.nodeName.toLowerCase();
    return l && (l === "input" && (t.type === "text" || t.type === "search" || t.type === "tel" || t.type === "url" || t.type === "password") || l === "textarea" || t.contentEditable === "true");
  }
  var Ch = Il && "documentMode" in document && 11 >= document.documentMode, Ea = null, Yi = null, ou = null, Gi = !1;
  function xs(t, l, e) {
    var a = e.window === e ? e.document : e.nodeType === 9 ? e : e.ownerDocument;
    Gi || Ea == null || Ea !== an(a) || (a = Ea, "selectionStart" in a && Bi(a) ? a = { start: a.selectionStart, end: a.selectionEnd } : (a = (a.ownerDocument && a.ownerDocument.defaultView || window).getSelection(), a = {
      anchorNode: a.anchorNode,
      anchorOffset: a.anchorOffset,
      focusNode: a.focusNode,
      focusOffset: a.focusOffset
    }), ou && ru(ou, a) || (ou = a, a = Wn(Yi, "onSelect"), 0 < a.length && (l = new sn(
      "onSelect",
      "select",
      null,
      l,
      e
    ), t.push({ event: l, listeners: a }), l.target = Ea)));
  }
  function ke(t, l) {
    var e = {};
    return e[t.toLowerCase()] = l.toLowerCase(), e["Webkit" + t] = "webkit" + l, e["Moz" + t] = "moz" + l, e;
  }
  var Aa = {
    animationend: ke("Animation", "AnimationEnd"),
    animationiteration: ke("Animation", "AnimationIteration"),
    animationstart: ke("Animation", "AnimationStart"),
    transitionrun: ke("Transition", "TransitionRun"),
    transitionstart: ke("Transition", "TransitionStart"),
    transitioncancel: ke("Transition", "TransitionCancel"),
    transitionend: ke("Transition", "TransitionEnd")
  }, Xi = {}, js = {};
  Il && (js = document.createElement("div").style, "AnimationEvent" in window || (delete Aa.animationend.animation, delete Aa.animationiteration.animation, delete Aa.animationstart.animation), "TransitionEvent" in window || delete Aa.transitionend.transition);
  function $e(t) {
    if (Xi[t]) return Xi[t];
    if (!Aa[t]) return t;
    var l = Aa[t], e;
    for (e in l)
      if (l.hasOwnProperty(e) && e in js)
        return Xi[t] = l[e];
    return t;
  }
  var Ts = $e("animationend"), Es = $e("animationiteration"), As = $e("animationstart"), Hh = $e("transitionrun"), qh = $e("transitionstart"), Bh = $e("transitioncancel"), zs = $e("transitionend"), Ms = /* @__PURE__ */ new Map(), Qi = "abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(
    " "
  );
  Qi.push("scrollEnd");
  function Gl(t, l) {
    Ms.set(t, l), Ke(l, [t]);
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
  function dn() {
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
  function hn(t, l, e, a) {
    Rl[za++] = t, Rl[za++] = l, Rl[za++] = e, Rl[za++] = a, Zi |= a, t.lanes |= a, t = t.alternate, t !== null && (t.lanes |= a);
  }
  function Li(t, l, e, a) {
    return hn(t, l, e, a), yn(t);
  }
  function Ma(t, l) {
    return hn(t, null, null, l), yn(t);
  }
  function Ds(t, l, e) {
    t.lanes |= e;
    var a = t.alternate;
    a !== null && (a.lanes |= e);
    for (var u = !1, n = t.return; n !== null; )
      n.childLanes |= e, a = n.alternate, a !== null && (a.childLanes |= e), n.tag === 22 && (t = n.stateNode, t === null || t._visibility & 1 || (u = !0)), t = n, n = n.return;
    return t.tag === 3 ? (n = t.stateNode, u && l !== null && (u = 31 - Pt(e), t = n.hiddenUpdates, a = t[u], a === null ? t[u] = [l] : a.push(l), l.lane = e | 536870912), n) : null;
  }
  function yn(t) {
    if (50 < qu)
      throw qu = 0, $c = null, Error(r(185));
    for (var l = t.return; l !== null; )
      t = l, l = t.return;
    return t.tag === 3 ? t.stateNode : null;
  }
  var Na = {};
  function Yh(t, l, e, a) {
    this.tag = t, this.key = e, this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null, this.index = 0, this.refCleanup = this.ref = null, this.pendingProps = l, this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null, this.mode = a, this.subtreeFlags = this.flags = 0, this.deletions = null, this.childLanes = this.lanes = 0, this.alternate = null;
  }
  function jl(t, l, e, a) {
    return new Yh(t, l, e, a);
  }
  function Vi(t) {
    return t = t.prototype, !(!t || !t.isReactComponent);
  }
  function te(t, l) {
    var e = t.alternate;
    return e === null ? (e = jl(
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
  function vn(t, l, e, a, u, n) {
    var i = 0;
    if (a = t, typeof t == "function") Vi(t) && (i = 1);
    else if (typeof t == "string")
      i = X0(
        t,
        e,
        w.current
      ) ? 26 : t === "html" || t === "head" || t === "body" ? 27 : 5;
    else
      t: switch (t) {
        case Ut:
          return t = jl(31, e, l, u), t.elementType = Ut, t.lanes = n, t;
        case B:
          return We(e.children, u, n, l);
        case yt:
          i = 8, u |= 24;
          break;
        case Nt:
          return t = jl(12, e, l, u | 2), t.elementType = Nt, t.lanes = n, t;
        case P:
          return t = jl(13, e, l, u), t.elementType = P, t.lanes = n, t;
        case at:
          return t = jl(19, e, l, u), t.elementType = at, t.lanes = n, t;
        default:
          if (typeof t == "object" && t !== null)
            switch (t.$$typeof) {
              case Wt:
              case Et:
                i = 10;
                break t;
              case pt:
                i = 9;
                break t;
              case Bt:
                i = 11;
                break t;
              case Ft:
                i = 14;
                break t;
              case V:
                i = 16, a = null;
                break t;
            }
          i = 29, e = Error(
            r(130, t === null ? "null" : typeof t, "")
          ), a = null;
      }
    return l = jl(i, e, l, u), l.elementType = t, l.type = a, l.lanes = n, l;
  }
  function We(t, l, e, a) {
    return t = jl(7, t, a, l), t.lanes = e, t;
  }
  function wi(t, l, e) {
    return t = jl(6, t, null, l), t.lanes = e, t;
  }
  function Ki(t, l, e) {
    return l = jl(
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
  var Da = [], Oa = 0, mn = null, gn = 0, _l = [], Ul = 0, Fe = null, le = 1, ee = "";
  function Pe(t, l) {
    Da[Oa++] = gn, Da[Oa++] = mn, mn = t, gn = l;
  }
  function Rs(t, l, e) {
    _l[Ul++] = le, _l[Ul++] = ee, _l[Ul++] = Fe, Fe = t;
    var a = le;
    t = ee;
    var u = 32 - Pt(a) - 1;
    a &= ~(1 << u), e += 1;
    var n = 32 - Pt(l) + u;
    if (30 < n) {
      var i = u - u % 5;
      n = (a & (1 << i) - 1).toString(32), a >>= i, u -= i, le = 1 << 32 - Pt(l) + u | e << u | a, ee = n + t;
    } else
      le = 1 << n | e << u | a, ee = t;
  }
  function Ji(t) {
    t.return !== null && (Pe(t, 1), Rs(t, 1, 0));
  }
  function ki(t) {
    for (; t === mn; )
      mn = Da[--Oa], Da[Oa] = null, gn = Da[--Oa], Da[Oa] = null;
    for (; t === Fe; )
      Fe = _l[--Ul], _l[Ul] = null, ee = _l[--Ul], _l[Ul] = null, le = _l[--Ul], _l[Ul] = null;
  }
  var cl = null, Rt = null, ht = !1, Ie = null, Vl = !1, $i = Error(r(519));
  function ta(t) {
    var l = Error(r(418, ""));
    throw yu(Ol(l, t)), $i;
  }
  function _s(t) {
    var l = t.stateNode, e = t.type, a = t.memoizedProps;
    switch (l[Yt] = t, l[al] = a, e) {
      case "dialog":
        nt("cancel", l), nt("close", l);
        break;
      case "iframe":
      case "object":
      case "embed":
        nt("load", l);
        break;
      case "video":
      case "audio":
        for (e = 0; e < Yu.length; e++)
          nt(Yu[e], l);
        break;
      case "source":
        nt("error", l);
        break;
      case "img":
      case "image":
      case "link":
        nt("error", l), nt("load", l);
        break;
      case "details":
        nt("toggle", l);
        break;
      case "input":
        nt("invalid", l), Jf(
          l,
          a.value,
          a.defaultValue,
          a.checked,
          a.defaultChecked,
          a.type,
          a.name,
          !0
        ), en(l);
        break;
      case "select":
        nt("invalid", l);
        break;
      case "textarea":
        nt("invalid", l), $f(l, a.value, a.defaultValue, a.children), en(l);
    }
    e = a.children, typeof e != "string" && typeof e != "number" && typeof e != "bigint" || l.textContent === "" + e || a.suppressHydrationWarning === !0 || Wo(l.textContent, e) ? (a.popover != null && (nt("beforetoggle", l), nt("toggle", l)), a.onScroll != null && nt("scroll", l), a.onScrollEnd != null && nt("scrollend", l), a.onClick != null && (l.onclick = Fn), l = !0) : l = !1, l || ta(t);
  }
  function Us(t) {
    for (cl = t.return; cl; )
      switch (cl.tag) {
        case 5:
        case 13:
          Vl = !1;
          return;
        case 27:
        case 3:
          Vl = !0;
          return;
        default:
          cl = cl.return;
      }
  }
  function du(t) {
    if (t !== cl) return !1;
    if (!ht) return Us(t), ht = !0, !1;
    var l = t.tag, e;
    if ((e = l !== 3 && l !== 27) && ((e = l === 5) && (e = t.type, e = !(e !== "form" && e !== "button") || df(t.type, t.memoizedProps)), e = !e), e && Rt && ta(t), Us(t), l === 13) {
      if (t = t.memoizedState, t = t !== null ? t.dehydrated : null, !t) throw Error(r(317));
      t: {
        for (t = t.nextSibling, l = 0; t; ) {
          if (t.nodeType === 8)
            if (e = t.data, e === "/$") {
              if (l === 0) {
                Rt = Ql(t.nextSibling);
                break t;
              }
              l--;
            } else
              e !== "$" && e !== "$!" && e !== "$?" || l++;
          t = t.nextSibling;
        }
        Rt = null;
      }
    } else
      l === 27 ? (l = Rt, Ce(t.type) ? (t = mf, mf = null, Rt = t) : Rt = l) : Rt = cl ? Ql(t.stateNode.nextSibling) : null;
    return !0;
  }
  function hu() {
    Rt = cl = null, ht = !1;
  }
  function Cs() {
    var t = Ie;
    return t !== null && (ml === null ? ml = t : ml.push.apply(
      ml,
      t
    ), Ie = null), t;
  }
  function yu(t) {
    Ie === null ? Ie = [t] : Ie.push(t);
  }
  var Wi = N(null), la = null, ae = null;
  function be(t, l, e) {
    q(Wi, l._currentValue), l._currentValue = e;
  }
  function ue(t) {
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
        if (i = u.return, i === null) throw Error(r(341));
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
  function vu(t, l, e, a) {
    t = null;
    for (var u = l, n = !1; u !== null; ) {
      if (!n) {
        if ((u.flags & 524288) !== 0) n = !0;
        else if ((u.flags & 262144) !== 0) break;
      }
      if (u.tag === 10) {
        var i = u.alternate;
        if (i === null) throw Error(r(387));
        if (i = i.memoizedProps, i !== null) {
          var f = u.type;
          xl(u.pendingProps.value, i.value) || (t !== null ? t.push(f) : t = [f]);
        }
      } else if (u === el.current) {
        if (i = u.alternate, i === null) throw Error(r(387));
        i.memoizedState.memoizedState !== u.memoizedState.memoizedState && (t !== null ? t.push(Vu) : t = [Vu]);
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
  function Sn(t) {
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
  function ea(t) {
    la = t, ae = null, t = t.dependencies, t !== null && (t.firstContext = null);
  }
  function ul(t) {
    return Hs(la, t);
  }
  function bn(t, l) {
    return la === null && ea(t), Hs(t, l);
  }
  function Hs(t, l) {
    var e = l._currentValue;
    if (l = { context: l, memoizedValue: e, next: null }, ae === null) {
      if (t === null) throw Error(r(308));
      ae = l, t.dependencies = { lanes: 0, firstContext: l }, t.flags |= 524288;
    } else ae = ae.next = l;
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
  }, Xh = o.unstable_scheduleCallback, Qh = o.unstable_NormalPriority, Vt = {
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
  function mu(t) {
    t.refCount--, t.refCount === 0 && Xh(Qh, function() {
      t.controller.abort();
    });
  }
  var gu = null, tc = 0, Ra = 0, _a = null;
  function Zh(t, l) {
    if (gu === null) {
      var e = gu = [];
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
    if (--tc === 0 && gu !== null) {
      _a !== null && (_a.status = "fulfilled");
      var t = gu;
      gu = null, Ra = 0, _a = null;
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
  var Bs = x.S;
  x.S = function(t, l) {
    typeof l == "object" && l !== null && typeof l.then == "function" && Zh(t, l), Bs !== null && Bs(t, l);
  };
  var aa = N(null);
  function lc() {
    var t = aa.current;
    return t !== null ? t : At.pooledCache;
  }
  function pn(t, l) {
    l === null ? q(aa, aa.current) : q(aa, l.pool);
  }
  function Ys() {
    var t = lc();
    return t === null ? null : { parent: Vt._currentValue, pool: t };
  }
  var Su = Error(r(460)), Gs = Error(r(474)), xn = Error(r(542)), ec = { then: function() {
  } };
  function Xs(t) {
    return t = t.status, t === "fulfilled" || t === "rejected";
  }
  function jn() {
  }
  function Qs(t, l, e) {
    switch (e = t[e], e === void 0 ? t.push(l) : e !== l && (l.then(jn, jn), l = e), l.status) {
      case "fulfilled":
        return l.value;
      case "rejected":
        throw t = l.reason, Ls(t), t;
      default:
        if (typeof l.status == "string") l.then(jn, jn);
        else {
          if (t = At, t !== null && 100 < t.shellSuspendCounter)
            throw Error(r(482));
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
        throw bu = l, Su;
    }
  }
  var bu = null;
  function Zs() {
    if (bu === null) throw Error(r(459));
    var t = bu;
    return bu = null, t;
  }
  function Ls(t) {
    if (t === Su || t === xn)
      throw Error(r(483));
  }
  var pe = !1;
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
  function xe(t) {
    return { lane: t, tag: 0, payload: null, callback: null, next: null };
  }
  function je(t, l, e) {
    var a = t.updateQueue;
    if (a === null) return null;
    if (a = a.shared, (vt & 2) !== 0) {
      var u = a.pending;
      return u === null ? l.next = l : (l.next = u.next, u.next = l), a.pending = l, l = yn(t), Ds(t, null, e), l;
    }
    return hn(t, a, l, e), yn(t);
  }
  function pu(t, l, e) {
    if (l = l.updateQueue, l !== null && (l = l.shared, (e & 4194048) !== 0)) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, rt(t, e);
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
  function ju(t, l, e, a) {
    ic = !1;
    var u = t.updateQueue;
    pe = !1;
    var n = u.firstBaseUpdate, i = u.lastBaseUpdate, f = u.shared.pending;
    if (f !== null) {
      u.shared.pending = null;
      var s = f, m = s.next;
      s.next = null, i === null ? n = m : i.next = m, i = s;
      var T = t.alternate;
      T !== null && (T = T.updateQueue, f = T.lastBaseUpdate, f !== i && (f === null ? T.firstBaseUpdate = m : f.next = m, T.lastBaseUpdate = s));
    }
    if (n !== null) {
      var z = u.baseState;
      i = 0, T = m = s = null, f = n;
      do {
        var g = f.lane & -536870913, S = g !== f.lane;
        if (S ? (st & g) === g : (a & g) === g) {
          g !== 0 && g === Ra && (ic = !0), T !== null && (T = T.next = {
            lane: 0,
            tag: f.tag,
            payload: f.payload,
            callback: null,
            next: null
          });
          t: {
            var W = t, J = f;
            g = l;
            var bt = e;
            switch (J.tag) {
              case 1:
                if (W = J.payload, typeof W == "function") {
                  z = W.call(bt, z, g);
                  break t;
                }
                z = W;
                break t;
              case 3:
                W.flags = W.flags & -65537 | 128;
              case 0:
                if (W = J.payload, g = typeof W == "function" ? W.call(bt, z, g) : W, g == null) break t;
                z = C({}, z, g);
                break t;
              case 2:
                pe = !0;
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
          }, T === null ? (m = T = S, s = z) : T = T.next = S, i |= g;
        if (f = f.next, f === null) {
          if (f = u.shared.pending, f === null)
            break;
          S = f, f = S.next, S.next = null, u.lastBaseUpdate = S, u.shared.pending = null;
        }
      } while (!0);
      T === null && (s = z), u.baseState = s, u.firstBaseUpdate = m, u.lastBaseUpdate = T, n === null && (u.shared.lanes = 0), Oe |= i, t.lanes = i, t.memoizedState = z;
    }
  }
  function Vs(t, l) {
    if (typeof t != "function")
      throw Error(r(191, t));
    t.call(l);
  }
  function ws(t, l) {
    var e = t.callbacks;
    if (e !== null)
      for (t.callbacks = null, t = 0; t < e.length; t++)
        Vs(e[t], l);
  }
  var Ua = N(null), Tn = N(0);
  function Ks(t, l) {
    t = oe, q(Tn, t), q(Ua, l), oe = t | l.baseLanes;
  }
  function cc() {
    q(Tn, oe), q(Ua, Ua.current);
  }
  function fc() {
    oe = Tn.current, Y(Ua), Y(Tn);
  }
  var Te = 0, lt = null, gt = null, Gt = null, En = !1, Ca = !1, ua = !1, An = 0, Tu = 0, Ha = null, Vh = 0;
  function Ct() {
    throw Error(r(321));
  }
  function sc(t, l) {
    if (l === null) return !1;
    for (var e = 0; e < l.length && e < t.length; e++)
      if (!xl(t[e], l[e])) return !1;
    return !0;
  }
  function rc(t, l, e, a, u, n) {
    return Te = n, lt = l, l.memoizedState = null, l.updateQueue = null, l.lanes = 0, x.H = t === null || t.memoizedState === null ? Dr : Or, ua = !1, n = e(a, u), ua = !1, Ca && (n = ks(
      l,
      e,
      a,
      u
    )), Js(t), n;
  }
  function Js(t) {
    x.H = Rn;
    var l = gt !== null && gt.next !== null;
    if (Te = 0, Gt = gt = lt = null, En = !1, Tu = 0, Ha = null, l) throw Error(r(300));
    t === null || Jt || (t = t.dependencies, t !== null && Sn(t) && (Jt = !0));
  }
  function ks(t, l, e, a) {
    lt = t;
    var u = 0;
    do {
      if (Ca && (Ha = null), Tu = 0, Ca = !1, 25 <= u) throw Error(r(301));
      if (u += 1, Gt = gt = null, t.updateQueue != null) {
        var n = t.updateQueue;
        n.lastEffect = null, n.events = null, n.stores = null, n.memoCache != null && (n.memoCache.index = 0);
      }
      x.H = Fh, n = l(e, a);
    } while (Ca);
    return n;
  }
  function wh() {
    var t = x.H, l = t.useState()[0];
    return l = typeof l.then == "function" ? Eu(l) : l, t = t.useState()[0], (gt !== null ? gt.memoizedState : null) !== t && (lt.flags |= 1024), l;
  }
  function oc() {
    var t = An !== 0;
    return An = 0, t;
  }
  function dc(t, l, e) {
    l.updateQueue = t.updateQueue, l.flags &= -2053, t.lanes &= ~e;
  }
  function hc(t) {
    if (En) {
      for (t = t.memoizedState; t !== null; ) {
        var l = t.queue;
        l !== null && (l.pending = null), t = t.next;
      }
      En = !1;
    }
    Te = 0, Gt = gt = lt = null, Ca = !1, Tu = An = 0, Ha = null;
  }
  function yl() {
    var t = {
      memoizedState: null,
      baseState: null,
      baseQueue: null,
      queue: null,
      next: null
    };
    return Gt === null ? lt.memoizedState = Gt = t : Gt = Gt.next = t, Gt;
  }
  function Xt() {
    if (gt === null) {
      var t = lt.alternate;
      t = t !== null ? t.memoizedState : null;
    } else t = gt.next;
    var l = Gt === null ? lt.memoizedState : Gt.next;
    if (l !== null)
      Gt = l, gt = t;
    else {
      if (t === null)
        throw lt.alternate === null ? Error(r(467)) : Error(r(310));
      gt = t, t = {
        memoizedState: gt.memoizedState,
        baseState: gt.baseState,
        baseQueue: gt.baseQueue,
        queue: gt.queue,
        next: null
      }, Gt === null ? lt.memoizedState = Gt = t : Gt = Gt.next = t;
    }
    return Gt;
  }
  function yc() {
    return { lastEffect: null, events: null, stores: null, memoCache: null };
  }
  function Eu(t) {
    var l = Tu;
    return Tu += 1, Ha === null && (Ha = []), t = Qs(Ha, t, l), l = lt, (Gt === null ? l.memoizedState : Gt.next) === null && (l = l.alternate, x.H = l === null || l.memoizedState === null ? Dr : Or), t;
  }
  function zn(t) {
    if (t !== null && typeof t == "object") {
      if (typeof t.then == "function") return Eu(t);
      if (t.$$typeof === Et) return ul(t);
    }
    throw Error(r(438, String(t)));
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
        e[a] = Bl;
    return l.index++, e;
  }
  function ne(t, l) {
    return typeof l == "function" ? l(t) : l;
  }
  function Mn(t) {
    var l = Xt();
    return mc(l, gt, t);
  }
  function mc(t, l, e) {
    var a = t.queue;
    if (a === null) throw Error(r(311));
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
      var f = i = null, s = null, m = l, T = !1;
      do {
        var z = m.lane & -536870913;
        if (z !== m.lane ? (st & z) === z : (Te & z) === z) {
          var g = m.revertLane;
          if (g === 0)
            s !== null && (s = s.next = {
              lane: 0,
              revertLane: 0,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }), z === Ra && (T = !0);
          else if ((Te & g) === g) {
            m = m.next, g === Ra && (T = !0);
            continue;
          } else
            z = {
              lane: 0,
              revertLane: m.revertLane,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }, s === null ? (f = s = z, i = n) : s = s.next = z, lt.lanes |= g, Oe |= g;
          z = m.action, ua && e(n, z), n = m.hasEagerState ? m.eagerState : e(n, z);
        } else
          g = {
            lane: z,
            revertLane: m.revertLane,
            action: m.action,
            hasEagerState: m.hasEagerState,
            eagerState: m.eagerState,
            next: null
          }, s === null ? (f = s = g, i = n) : s = s.next = g, lt.lanes |= z, Oe |= z;
        m = m.next;
      } while (m !== null && m !== l);
      if (s === null ? i = n : s.next = f, !xl(n, t.memoizedState) && (Jt = !0, T && (e = _a, e !== null)))
        throw e;
      t.memoizedState = n, t.baseState = i, t.baseQueue = s, a.lastRenderedState = n;
    }
    return u === null && (a.lanes = 0), [t.memoizedState, a.dispatch];
  }
  function gc(t) {
    var l = Xt(), e = l.queue;
    if (e === null) throw Error(r(311));
    e.lastRenderedReducer = t;
    var a = e.dispatch, u = e.pending, n = l.memoizedState;
    if (u !== null) {
      e.pending = null;
      var i = u = u.next;
      do
        n = t(n, i.action), i = i.next;
      while (i !== u);
      xl(n, l.memoizedState) || (Jt = !0), l.memoizedState = n, l.baseQueue === null && (l.baseState = n), e.lastRenderedState = n;
    }
    return [n, a];
  }
  function $s(t, l, e) {
    var a = lt, u = Xt(), n = ht;
    if (n) {
      if (e === void 0) throw Error(r(407));
      e = e();
    } else e = l();
    var i = !xl(
      (gt || u).memoizedState,
      e
    );
    i && (u.memoizedState = e, Jt = !0), u = u.queue;
    var f = Ps.bind(null, a, u, t);
    if (Au(2048, 8, f, [t]), u.getSnapshot !== l || i || Gt !== null && Gt.memoizedState.tag & 1) {
      if (a.flags |= 2048, qa(
        9,
        Nn(),
        Fs.bind(
          null,
          a,
          u,
          e,
          l
        ),
        null
      ), At === null) throw Error(r(349));
      n || (Te & 124) !== 0 || Ws(a, l, e);
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
    l !== null && Ml(l, t, 2);
  }
  function Sc(t) {
    var l = yl();
    if (typeof t == "function") {
      var e = t;
      if (t = e(), ua) {
        Dt(!0);
        try {
          e();
        } finally {
          Dt(!1);
        }
      }
    }
    return l.memoizedState = l.baseState = t, l.queue = {
      pending: null,
      lanes: 0,
      dispatch: null,
      lastRenderedReducer: ne,
      lastRenderedState: t
    }, l;
  }
  function lr(t, l, e, a) {
    return t.baseState = e, mc(
      t,
      gt,
      typeof a == "function" ? a : ne
    );
  }
  function Kh(t, l, e, a, u) {
    if (On(t)) throw Error(r(485));
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
      x.T !== null ? e(!0) : n.isTransition = !1, a(n), e = l.pending, e === null ? (n.next = l.pending = n, er(l, n)) : (n.next = e.next, l.pending = e.next = n);
    }
  }
  function er(t, l) {
    var e = l.action, a = l.payload, u = t.state;
    if (l.isTransition) {
      var n = x.T, i = {};
      x.T = i;
      try {
        var f = e(u, a), s = x.S;
        s !== null && s(i, f), ar(t, l, f);
      } catch (m) {
        bc(t, l, m);
      } finally {
        x.T = n;
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
    if (ht) {
      var e = At.formState;
      if (e !== null) {
        t: {
          var a = lt;
          if (ht) {
            if (Rt) {
              l: {
                for (var u = Rt, n = Vl; u.nodeType !== 8; ) {
                  if (!n) {
                    u = null;
                    break l;
                  }
                  if (u = Ql(
                    u.nextSibling
                  ), u === null) {
                    u = null;
                    break l;
                  }
                }
                n = u.data, u = n === "F!" || n === "F" ? u : null;
              }
              if (u) {
                Rt = Ql(
                  u.nextSibling
                ), a = u.data === "F!";
                break t;
              }
            }
            ta(a);
          }
          a = !1;
        }
        a && (l = e[0]);
      }
    }
    return e = yl(), e.memoizedState = e.baseState = l, a = {
      pending: null,
      lanes: 0,
      dispatch: null,
      lastRenderedReducer: ir,
      lastRenderedState: l
    }, e.queue = a, e = zr.bind(
      null,
      lt,
      a
    ), a.dispatch = e, a = Sc(!1), n = Ec.bind(
      null,
      lt,
      !1,
      a.queue
    ), a = yl(), u = {
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
    var l = Xt();
    return sr(l, gt, t);
  }
  function sr(t, l, e) {
    if (l = mc(
      t,
      l,
      ir
    )[0], t = Mn(ne)[0], typeof l == "object" && l !== null && typeof l.then == "function")
      try {
        var a = Eu(l);
      } catch (i) {
        throw i === Su ? xn : i;
      }
    else a = l;
    l = Xt();
    var u = l.queue, n = u.dispatch;
    return e !== l.memoizedState && (lt.flags |= 2048, qa(
      9,
      Nn(),
      Jh.bind(null, u, e),
      null
    )), [a, n, t];
  }
  function Jh(t, l) {
    t.action = l;
  }
  function rr(t) {
    var l = Xt(), e = gt;
    if (e !== null)
      return sr(l, e, t);
    Xt(), l = l.memoizedState, e = Xt();
    var a = e.queue.dispatch;
    return e.memoizedState = t, [l, a, !1];
  }
  function qa(t, l, e, a) {
    return t = { tag: t, create: e, deps: a, inst: l, next: null }, l = lt.updateQueue, l === null && (l = yc(), lt.updateQueue = l), e = l.lastEffect, e === null ? l.lastEffect = t.next = t : (a = e.next, e.next = t, t.next = a, l.lastEffect = t), t;
  }
  function Nn() {
    return { destroy: void 0, resource: void 0 };
  }
  function or() {
    return Xt().memoizedState;
  }
  function Dn(t, l, e, a) {
    var u = yl();
    a = a === void 0 ? null : a, lt.flags |= t, u.memoizedState = qa(
      1 | l,
      Nn(),
      e,
      a
    );
  }
  function Au(t, l, e, a) {
    var u = Xt();
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
    Dn(8390656, 8, t, l);
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
  function pc() {
  }
  function Sr(t, l) {
    var e = Xt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    return l !== null && sc(l, a[1]) ? a[0] : (e.memoizedState = [t, l], t);
  }
  function br(t, l) {
    var e = Xt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    if (l !== null && sc(l, a[1]))
      return a[0];
    if (a = t(), ua) {
      Dt(!0);
      try {
        t();
      } finally {
        Dt(!1);
      }
    }
    return e.memoizedState = [a, l], a;
  }
  function xc(t, l, e) {
    return e === void 0 || (Te & 1073741824) !== 0 ? t.memoizedState = l : (t.memoizedState = e, t = To(), lt.lanes |= t, Oe |= t, e);
  }
  function pr(t, l, e, a) {
    return xl(e, l) ? e : Ua.current !== null ? (t = xc(t, e, a), xl(t, l) || (Jt = !0), t) : (Te & 42) === 0 ? (Jt = !0, t.memoizedState = e) : (t = To(), lt.lanes |= t, Oe |= t, l);
  }
  function xr(t, l, e, a, u) {
    var n = H.p;
    H.p = n !== 0 && 8 > n ? n : 8;
    var i = x.T, f = {};
    x.T = f, Ec(t, !1, l, e);
    try {
      var s = u(), m = x.S;
      if (m !== null && m(f, s), s !== null && typeof s == "object" && typeof s.then == "function") {
        var T = Lh(
          s,
          a
        );
        zu(
          t,
          l,
          T,
          zl(t)
        );
      } else
        zu(
          t,
          l,
          a,
          zl(t)
        );
    } catch (z) {
      zu(
        t,
        l,
        { then: function() {
        }, status: "rejected", reason: z },
        zl()
      );
    } finally {
      H.p = n, x.T = i;
    }
  }
  function kh() {
  }
  function jc(t, l, e, a) {
    if (t.tag !== 5) throw Error(r(476));
    var u = jr(t).queue;
    xr(
      t,
      u,
      l,
      K,
      e === null ? kh : function() {
        return Tr(t), e(a);
      }
    );
  }
  function jr(t) {
    var l = t.memoizedState;
    if (l !== null) return l;
    l = {
      memoizedState: K,
      baseState: K,
      baseQueue: null,
      queue: {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: ne,
        lastRenderedState: K
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
        lastRenderedReducer: ne,
        lastRenderedState: e
      },
      next: null
    }, t.memoizedState = l, t = t.alternate, t !== null && (t.memoizedState = l), l;
  }
  function Tr(t) {
    var l = jr(t).next.queue;
    zu(t, l, {}, zl());
  }
  function Tc() {
    return ul(Vu);
  }
  function Er() {
    return Xt().memoizedState;
  }
  function Ar() {
    return Xt().memoizedState;
  }
  function $h(t) {
    for (var l = t.return; l !== null; ) {
      switch (l.tag) {
        case 24:
        case 3:
          var e = zl();
          t = xe(e);
          var a = je(l, t, e);
          a !== null && (Ml(a, l, e), pu(a, l, e)), l = { cache: Ii() }, t.payload = l;
          return;
      }
      l = l.return;
    }
  }
  function Wh(t, l, e) {
    var a = zl();
    e = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, On(t) ? Mr(l, e) : (e = Li(t, l, e, a), e !== null && (Ml(e, t, a), Nr(e, l, a)));
  }
  function zr(t, l, e) {
    var a = zl();
    zu(t, l, e, a);
  }
  function zu(t, l, e, a) {
    var u = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    };
    if (On(t)) Mr(l, u);
    else {
      var n = t.alternate;
      if (t.lanes === 0 && (n === null || n.lanes === 0) && (n = l.lastRenderedReducer, n !== null))
        try {
          var i = l.lastRenderedState, f = n(i, e);
          if (u.hasEagerState = !0, u.eagerState = f, xl(f, i))
            return hn(t, l, u, 0), At === null && dn(), !1;
        } catch {
        }
      if (e = Li(t, l, u, a), e !== null)
        return Ml(e, t, a), Nr(e, l, a), !0;
    }
    return !1;
  }
  function Ec(t, l, e, a) {
    if (a = {
      lane: 2,
      revertLane: ef(),
      action: a,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, On(t)) {
      if (l) throw Error(r(479));
    } else
      l = Li(
        t,
        e,
        a,
        2
      ), l !== null && Ml(l, t, 2);
  }
  function On(t) {
    var l = t.alternate;
    return t === lt || l !== null && l === lt;
  }
  function Mr(t, l) {
    Ca = En = !0;
    var e = t.pending;
    e === null ? l.next = l : (l.next = e.next, e.next = l), t.pending = l;
  }
  function Nr(t, l, e) {
    if ((e & 4194048) !== 0) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, rt(t, e);
    }
  }
  var Rn = {
    readContext: ul,
    use: zn,
    useCallback: Ct,
    useContext: Ct,
    useEffect: Ct,
    useImperativeHandle: Ct,
    useLayoutEffect: Ct,
    useInsertionEffect: Ct,
    useMemo: Ct,
    useReducer: Ct,
    useRef: Ct,
    useState: Ct,
    useDebugValue: Ct,
    useDeferredValue: Ct,
    useTransition: Ct,
    useSyncExternalStore: Ct,
    useId: Ct,
    useHostTransitionStatus: Ct,
    useFormState: Ct,
    useActionState: Ct,
    useOptimistic: Ct,
    useMemoCache: Ct,
    useCacheRefresh: Ct
  }, Dr = {
    readContext: ul,
    use: zn,
    useCallback: function(t, l) {
      return yl().memoizedState = [
        t,
        l === void 0 ? null : l
      ], t;
    },
    useContext: ul,
    useEffect: dr,
    useImperativeHandle: function(t, l, e) {
      e = e != null ? e.concat([t]) : null, Dn(
        4194308,
        4,
        mr.bind(null, l, t),
        e
      );
    },
    useLayoutEffect: function(t, l) {
      return Dn(4194308, 4, t, l);
    },
    useInsertionEffect: function(t, l) {
      Dn(4, 2, t, l);
    },
    useMemo: function(t, l) {
      var e = yl();
      l = l === void 0 ? null : l;
      var a = t();
      if (ua) {
        Dt(!0);
        try {
          t();
        } finally {
          Dt(!1);
        }
      }
      return e.memoizedState = [a, l], a;
    },
    useReducer: function(t, l, e) {
      var a = yl();
      if (e !== void 0) {
        var u = e(l);
        if (ua) {
          Dt(!0);
          try {
            e(l);
          } finally {
            Dt(!1);
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
      var l = yl();
      return t = { current: t }, l.memoizedState = t;
    },
    useState: function(t) {
      t = Sc(t);
      var l = t.queue, e = zr.bind(null, lt, l);
      return l.dispatch = e, [t.memoizedState, e];
    },
    useDebugValue: pc,
    useDeferredValue: function(t, l) {
      var e = yl();
      return xc(e, t, l);
    },
    useTransition: function() {
      var t = Sc(!1);
      return t = xr.bind(
        null,
        lt,
        t.queue,
        !0,
        !1
      ), yl().memoizedState = t, [!1, t];
    },
    useSyncExternalStore: function(t, l, e) {
      var a = lt, u = yl();
      if (ht) {
        if (e === void 0)
          throw Error(r(407));
        e = e();
      } else {
        if (e = l(), At === null)
          throw Error(r(349));
        (st & 124) !== 0 || Ws(a, l, e);
      }
      u.memoizedState = e;
      var n = { value: e, getSnapshot: l };
      return u.queue = n, dr(Ps.bind(null, a, n, t), [
        t
      ]), a.flags |= 2048, qa(
        9,
        Nn(),
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
      var t = yl(), l = At.identifierPrefix;
      if (ht) {
        var e = ee, a = le;
        e = (a & ~(1 << 32 - Pt(a) - 1)).toString(32) + e, l = "«" + l + "R" + e, e = An++, 0 < e && (l += "H" + e.toString(32)), l += "»";
      } else
        e = Vh++, l = "«" + l + "r" + e.toString(32) + "»";
      return t.memoizedState = l;
    },
    useHostTransitionStatus: Tc,
    useFormState: cr,
    useActionState: cr,
    useOptimistic: function(t) {
      var l = yl();
      l.memoizedState = l.baseState = t;
      var e = {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: null,
        lastRenderedState: null
      };
      return l.queue = e, l = Ec.bind(
        null,
        lt,
        !0,
        e
      ), e.dispatch = l, [t, l];
    },
    useMemoCache: vc,
    useCacheRefresh: function() {
      return yl().memoizedState = $h.bind(
        null,
        lt
      );
    }
  }, Or = {
    readContext: ul,
    use: zn,
    useCallback: Sr,
    useContext: ul,
    useEffect: hr,
    useImperativeHandle: gr,
    useInsertionEffect: yr,
    useLayoutEffect: vr,
    useMemo: br,
    useReducer: Mn,
    useRef: or,
    useState: function() {
      return Mn(ne);
    },
    useDebugValue: pc,
    useDeferredValue: function(t, l) {
      var e = Xt();
      return pr(
        e,
        gt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = Mn(ne)[0], l = Xt().memoizedState;
      return [
        typeof t == "boolean" ? t : Eu(t),
        l
      ];
    },
    useSyncExternalStore: $s,
    useId: Er,
    useHostTransitionStatus: Tc,
    useFormState: fr,
    useActionState: fr,
    useOptimistic: function(t, l) {
      var e = Xt();
      return lr(e, gt, t, l);
    },
    useMemoCache: vc,
    useCacheRefresh: Ar
  }, Fh = {
    readContext: ul,
    use: zn,
    useCallback: Sr,
    useContext: ul,
    useEffect: hr,
    useImperativeHandle: gr,
    useInsertionEffect: yr,
    useLayoutEffect: vr,
    useMemo: br,
    useReducer: gc,
    useRef: or,
    useState: function() {
      return gc(ne);
    },
    useDebugValue: pc,
    useDeferredValue: function(t, l) {
      var e = Xt();
      return gt === null ? xc(e, t, l) : pr(
        e,
        gt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = gc(ne)[0], l = Xt().memoizedState;
      return [
        typeof t == "boolean" ? t : Eu(t),
        l
      ];
    },
    useSyncExternalStore: $s,
    useId: Er,
    useHostTransitionStatus: Tc,
    useFormState: rr,
    useActionState: rr,
    useOptimistic: function(t, l) {
      var e = Xt();
      return gt !== null ? lr(e, gt, t, l) : (e.baseState = t, [t, e.queue.dispatch]);
    },
    useMemoCache: vc,
    useCacheRefresh: Ar
  }, Ba = null, Mu = 0;
  function _n(t) {
    var l = Mu;
    return Mu += 1, Ba === null && (Ba = []), Qs(Ba, t, l);
  }
  function Nu(t, l) {
    l = l.props.ref, t.ref = l !== void 0 ? l : null;
  }
  function Un(t, l) {
    throw l.$$typeof === G ? Error(r(525)) : (t = Object.prototype.toString.call(l), Error(
      r(
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
      return y = te(y, h), y.index = 0, y.sibling = null, y;
    }
    function n(y, h, v) {
      return y.index = v, t ? (v = y.alternate, v !== null ? (v = v.index, v < h ? (y.flags |= 67108866, h) : v) : (y.flags |= 67108866, h)) : (y.flags |= 1048576, h);
    }
    function i(y) {
      return t && y.alternate === null && (y.flags |= 67108866), y;
    }
    function f(y, h, v, A) {
      return h === null || h.tag !== 6 ? (h = wi(v, y.mode, A), h.return = y, h) : (h = u(h, v), h.return = y, h);
    }
    function s(y, h, v, A) {
      var Q = v.type;
      return Q === B ? T(
        y,
        h,
        v.props.children,
        A,
        v.key
      ) : h !== null && (h.elementType === Q || typeof Q == "object" && Q !== null && Q.$$typeof === V && Rr(Q) === h.type) ? (h = u(h, v.props), Nu(h, v), h.return = y, h) : (h = vn(
        v.type,
        v.key,
        v.props,
        null,
        y.mode,
        A
      ), Nu(h, v), h.return = y, h);
    }
    function m(y, h, v, A) {
      return h === null || h.tag !== 4 || h.stateNode.containerInfo !== v.containerInfo || h.stateNode.implementation !== v.implementation ? (h = Ki(v, y.mode, A), h.return = y, h) : (h = u(h, v.children || []), h.return = y, h);
    }
    function T(y, h, v, A, Q) {
      return h === null || h.tag !== 7 ? (h = We(
        v,
        y.mode,
        A,
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
          case Z:
            return v = vn(
              h.type,
              h.key,
              h.props,
              null,
              y.mode,
              v
            ), Nu(v, h), v.return = y, v;
          case tt:
            return h = Ki(
              h,
              y.mode,
              v
            ), h.return = y, h;
          case V:
            var A = h._init;
            return h = A(h._payload), z(y, h, v);
        }
        if (Zt(h) || Qt(h))
          return h = We(
            h,
            y.mode,
            v,
            null
          ), h.return = y, h;
        if (typeof h.then == "function")
          return z(y, _n(h), v);
        if (h.$$typeof === Et)
          return z(
            y,
            bn(y, h),
            v
          );
        Un(y, h);
      }
      return null;
    }
    function g(y, h, v, A) {
      var Q = h !== null ? h.key : null;
      if (typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint")
        return Q !== null ? null : f(y, h, "" + v, A);
      if (typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case Z:
            return v.key === Q ? s(y, h, v, A) : null;
          case tt:
            return v.key === Q ? m(y, h, v, A) : null;
          case V:
            return Q = v._init, v = Q(v._payload), g(y, h, v, A);
        }
        if (Zt(v) || Qt(v))
          return Q !== null ? null : T(y, h, v, A, null);
        if (typeof v.then == "function")
          return g(
            y,
            h,
            _n(v),
            A
          );
        if (v.$$typeof === Et)
          return g(
            y,
            h,
            bn(y, v),
            A
          );
        Un(y, v);
      }
      return null;
    }
    function S(y, h, v, A, Q) {
      if (typeof A == "string" && A !== "" || typeof A == "number" || typeof A == "bigint")
        return y = y.get(v) || null, f(h, y, "" + A, Q);
      if (typeof A == "object" && A !== null) {
        switch (A.$$typeof) {
          case Z:
            return y = y.get(
              A.key === null ? v : A.key
            ) || null, s(h, y, A, Q);
          case tt:
            return y = y.get(
              A.key === null ? v : A.key
            ) || null, m(h, y, A, Q);
          case V:
            var et = A._init;
            return A = et(A._payload), S(
              y,
              h,
              v,
              A,
              Q
            );
        }
        if (Zt(A) || Qt(A))
          return y = y.get(v) || null, T(h, y, A, Q, null);
        if (typeof A.then == "function")
          return S(
            y,
            h,
            v,
            _n(A),
            Q
          );
        if (A.$$typeof === Et)
          return S(
            y,
            h,
            v,
            bn(h, A),
            Q
          );
        Un(h, A);
      }
      return null;
    }
    function W(y, h, v, A) {
      for (var Q = null, et = null, L = h, k = h = 0, $t = null; L !== null && k < v.length; k++) {
        L.index > k ? ($t = L, L = null) : $t = L.sibling;
        var ot = g(
          y,
          L,
          v[k],
          A
        );
        if (ot === null) {
          L === null && (L = $t);
          break;
        }
        t && L && ot.alternate === null && l(y, L), h = n(ot, h, k), et === null ? Q = ot : et.sibling = ot, et = ot, L = $t;
      }
      if (k === v.length)
        return e(y, L), ht && Pe(y, k), Q;
      if (L === null) {
        for (; k < v.length; k++)
          L = z(y, v[k], A), L !== null && (h = n(
            L,
            h,
            k
          ), et === null ? Q = L : et.sibling = L, et = L);
        return ht && Pe(y, k), Q;
      }
      for (L = a(L); k < v.length; k++)
        $t = S(
          L,
          y,
          k,
          v[k],
          A
        ), $t !== null && (t && $t.alternate !== null && L.delete(
          $t.key === null ? k : $t.key
        ), h = n(
          $t,
          h,
          k
        ), et === null ? Q = $t : et.sibling = $t, et = $t);
      return t && L.forEach(function(Ge) {
        return l(y, Ge);
      }), ht && Pe(y, k), Q;
    }
    function J(y, h, v, A) {
      if (v == null) throw Error(r(151));
      for (var Q = null, et = null, L = h, k = h = 0, $t = null, ot = v.next(); L !== null && !ot.done; k++, ot = v.next()) {
        L.index > k ? ($t = L, L = null) : $t = L.sibling;
        var Ge = g(y, L, ot.value, A);
        if (Ge === null) {
          L === null && (L = $t);
          break;
        }
        t && L && Ge.alternate === null && l(y, L), h = n(Ge, h, k), et === null ? Q = Ge : et.sibling = Ge, et = Ge, L = $t;
      }
      if (ot.done)
        return e(y, L), ht && Pe(y, k), Q;
      if (L === null) {
        for (; !ot.done; k++, ot = v.next())
          ot = z(y, ot.value, A), ot !== null && (h = n(ot, h, k), et === null ? Q = ot : et.sibling = ot, et = ot);
        return ht && Pe(y, k), Q;
      }
      for (L = a(L); !ot.done; k++, ot = v.next())
        ot = S(L, y, k, ot.value, A), ot !== null && (t && ot.alternate !== null && L.delete(ot.key === null ? k : ot.key), h = n(ot, h, k), et === null ? Q = ot : et.sibling = ot, et = ot);
      return t && L.forEach(function(P0) {
        return l(y, P0);
      }), ht && Pe(y, k), Q;
    }
    function bt(y, h, v, A) {
      if (typeof v == "object" && v !== null && v.type === B && v.key === null && (v = v.props.children), typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case Z:
            t: {
              for (var Q = v.key; h !== null; ) {
                if (h.key === Q) {
                  if (Q = v.type, Q === B) {
                    if (h.tag === 7) {
                      e(
                        y,
                        h.sibling
                      ), A = u(
                        h,
                        v.props.children
                      ), A.return = y, y = A;
                      break t;
                    }
                  } else if (h.elementType === Q || typeof Q == "object" && Q !== null && Q.$$typeof === V && Rr(Q) === h.type) {
                    e(
                      y,
                      h.sibling
                    ), A = u(h, v.props), Nu(A, v), A.return = y, y = A;
                    break t;
                  }
                  e(y, h);
                  break;
                } else l(y, h);
                h = h.sibling;
              }
              v.type === B ? (A = We(
                v.props.children,
                y.mode,
                A,
                v.key
              ), A.return = y, y = A) : (A = vn(
                v.type,
                v.key,
                v.props,
                null,
                y.mode,
                A
              ), Nu(A, v), A.return = y, y = A);
            }
            return i(y);
          case tt:
            t: {
              for (Q = v.key; h !== null; ) {
                if (h.key === Q)
                  if (h.tag === 4 && h.stateNode.containerInfo === v.containerInfo && h.stateNode.implementation === v.implementation) {
                    e(
                      y,
                      h.sibling
                    ), A = u(h, v.children || []), A.return = y, y = A;
                    break t;
                  } else {
                    e(y, h);
                    break;
                  }
                else l(y, h);
                h = h.sibling;
              }
              A = Ki(v, y.mode, A), A.return = y, y = A;
            }
            return i(y);
          case V:
            return Q = v._init, v = Q(v._payload), bt(
              y,
              h,
              v,
              A
            );
        }
        if (Zt(v))
          return W(
            y,
            h,
            v,
            A
          );
        if (Qt(v)) {
          if (Q = Qt(v), typeof Q != "function") throw Error(r(150));
          return v = Q.call(v), J(
            y,
            h,
            v,
            A
          );
        }
        if (typeof v.then == "function")
          return bt(
            y,
            h,
            _n(v),
            A
          );
        if (v.$$typeof === Et)
          return bt(
            y,
            h,
            bn(y, v),
            A
          );
        Un(y, v);
      }
      return typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint" ? (v = "" + v, h !== null && h.tag === 6 ? (e(y, h.sibling), A = u(h, v), A.return = y, y = A) : (e(y, h), A = wi(v, y.mode, A), A.return = y, y = A), i(y)) : e(y, h);
    }
    return function(y, h, v, A) {
      try {
        Mu = 0;
        var Q = bt(
          y,
          h,
          v,
          A
        );
        return Ba = null, Q;
      } catch (L) {
        if (L === Su || L === xn) throw L;
        var et = jl(29, L, null, y.mode);
        return et.lanes = A, et.return = y, et;
      }
    };
  }
  var Ya = _r(!0), Ur = _r(!1), Cl = N(null), wl = null;
  function Ee(t) {
    var l = t.alternate;
    q(wt, wt.current & 1), q(Cl, t), wl === null && (l === null || Ua.current !== null || l.memoizedState !== null) && (wl = t);
  }
  function Cr(t) {
    if (t.tag === 22) {
      if (q(wt, wt.current), q(Cl, t), wl === null) {
        var l = t.alternate;
        l !== null && l.memoizedState !== null && (wl = t);
      }
    } else Ae();
  }
  function Ae() {
    q(wt, wt.current), q(Cl, Cl.current);
  }
  function ie(t) {
    Y(Cl), wl === t && (wl = null), Y(wt);
  }
  var wt = N(0);
  function Cn(t) {
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
  function Ac(t, l, e, a) {
    l = t.memoizedState, e = e(a, l), e = e == null ? l : C({}, l, e), t.memoizedState = e, t.lanes === 0 && (t.updateQueue.baseState = e);
  }
  var zc = {
    enqueueSetState: function(t, l, e) {
      t = t._reactInternals;
      var a = zl(), u = xe(a);
      u.payload = l, e != null && (u.callback = e), l = je(t, u, a), l !== null && (Ml(l, t, a), pu(l, t, a));
    },
    enqueueReplaceState: function(t, l, e) {
      t = t._reactInternals;
      var a = zl(), u = xe(a);
      u.tag = 1, u.payload = l, e != null && (u.callback = e), l = je(t, u, a), l !== null && (Ml(l, t, a), pu(l, t, a));
    },
    enqueueForceUpdate: function(t, l) {
      t = t._reactInternals;
      var e = zl(), a = xe(e);
      a.tag = 2, l != null && (a.callback = l), l = je(t, a, e), l !== null && (Ml(l, t, e), pu(l, t, e));
    }
  };
  function Hr(t, l, e, a, u, n, i) {
    return t = t.stateNode, typeof t.shouldComponentUpdate == "function" ? t.shouldComponentUpdate(a, n, i) : l.prototype && l.prototype.isPureReactComponent ? !ru(e, a) || !ru(u, n) : !0;
  }
  function qr(t, l, e, a) {
    t = l.state, typeof l.componentWillReceiveProps == "function" && l.componentWillReceiveProps(e, a), typeof l.UNSAFE_componentWillReceiveProps == "function" && l.UNSAFE_componentWillReceiveProps(e, a), l.state !== t && zc.enqueueReplaceState(l, l.state, null);
  }
  function na(t, l) {
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
  var Hn = typeof reportError == "function" ? reportError : function(t) {
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
    Hn(t);
  }
  function Yr(t) {
    console.error(t);
  }
  function Gr(t) {
    Hn(t);
  }
  function qn(t, l) {
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
    return e = xe(e), e.tag = 3, e.payload = { element: null }, e.callback = function() {
      qn(t, l);
    }, e;
  }
  function Qr(t) {
    return t = xe(t), t.tag = 3, t;
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
      Xr(l, e, a), typeof u != "function" && (Re === null ? Re = /* @__PURE__ */ new Set([this]) : Re.add(this));
      var f = a.stack;
      this.componentDidCatch(a.value, {
        componentStack: f !== null ? f : ""
      });
    });
  }
  function Ph(t, l, e, a, u) {
    if (e.flags |= 32768, a !== null && typeof a == "object" && typeof a.then == "function") {
      if (l = e.alternate, l !== null && vu(
        l,
        e,
        u,
        !0
      ), e = Cl.current, e !== null) {
        switch (e.tag) {
          case 13:
            return wl === null ? Fc() : e.alternate === null && _t === 0 && (_t = 3), e.flags &= -257, e.flags |= 65536, e.lanes = u, a === ec ? e.flags |= 16384 : (l = e.updateQueue, l === null ? e.updateQueue = /* @__PURE__ */ new Set([a]) : l.add(a), Ic(t, a, u)), !1;
          case 22:
            return e.flags |= 65536, a === ec ? e.flags |= 16384 : (l = e.updateQueue, l === null ? (l = {
              transitions: null,
              markerInstances: null,
              retryQueue: /* @__PURE__ */ new Set([a])
            }, e.updateQueue = l) : (e = l.retryQueue, e === null ? l.retryQueue = /* @__PURE__ */ new Set([a]) : e.add(a)), Ic(t, a, u)), !1;
        }
        throw Error(r(435, e.tag));
      }
      return Ic(t, a, u), Fc(), !1;
    }
    if (ht)
      return l = Cl.current, l !== null ? ((l.flags & 65536) === 0 && (l.flags |= 256), l.flags |= 65536, l.lanes = u, a !== $i && (t = Error(r(422), { cause: a }), yu(Ol(t, e)))) : (a !== $i && (l = Error(r(423), {
        cause: a
      }), yu(
        Ol(l, e)
      )), t = t.current.alternate, t.flags |= 65536, u &= -u, t.lanes |= u, a = Ol(a, e), u = Mc(
        t.stateNode,
        a,
        u
      ), nc(t, u), _t !== 4 && (_t = 2)), !1;
    var n = Error(r(520), { cause: a });
    if (n = Ol(n, e), Hu === null ? Hu = [n] : Hu.push(n), _t !== 4 && (_t = 2), l === null) return !0;
    a = Ol(a, e), e = l;
    do {
      switch (e.tag) {
        case 3:
          return e.flags |= 65536, t = u & -u, e.lanes |= t, t = Mc(e.stateNode, a, t), nc(e, t), !1;
        case 1:
          if (l = e.type, n = e.stateNode, (e.flags & 128) === 0 && (typeof l.getDerivedStateFromError == "function" || n !== null && typeof n.componentDidCatch == "function" && (Re === null || !Re.has(n))))
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
  var Lr = Error(r(461)), Jt = !1;
  function It(t, l, e, a) {
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
    return ea(l), a = rc(
      t,
      l,
      e,
      i,
      n,
      u
    ), f = oc(), t !== null && !Jt ? (dc(t, l, u), ce(t, l, u)) : (ht && f && Ji(l), l.flags |= 1, It(t, l, a, u), l.child);
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
      )) : (t = vn(
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
      if (e = e.compare, e = e !== null ? e : ru, e(i, a) && t.ref === l.ref)
        return ce(t, l, u);
    }
    return l.flags |= 1, t = te(n, a), t.ref = l.ref, t.return = l, l.child = t;
  }
  function Kr(t, l, e, a, u) {
    if (t !== null) {
      var n = t.memoizedProps;
      if (ru(n, a) && t.ref === l.ref)
        if (Jt = !1, l.pendingProps = a = n, Hc(t, u))
          (t.flags & 131072) !== 0 && (Jt = !0);
        else
          return l.lanes = t.lanes, ce(t, l, u);
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
        l.memoizedState = { baseLanes: 0, cachePool: null }, t !== null && pn(
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
      n !== null ? (pn(l, n.cachePool), Ks(l, n), Ae(), l.memoizedState = null) : (t !== null && pn(l, null), cc(), Ae());
    return It(t, l, u, e), l.child;
  }
  function kr(t, l, e, a) {
    var u = lc();
    return u = u === null ? null : { parent: Vt._currentValue, pool: u }, l.memoizedState = {
      baseLanes: e,
      cachePool: u
    }, t !== null && pn(l, null), cc(), Cr(l), t !== null && vu(t, l, a, !0), null;
  }
  function Bn(t, l) {
    var e = l.ref;
    if (e === null)
      t !== null && t.ref !== null && (l.flags |= 4194816);
    else {
      if (typeof e != "function" && typeof e != "object")
        throw Error(r(284));
      (t === null || t.ref !== e) && (l.flags |= 4194816);
    }
  }
  function Nc(t, l, e, a, u) {
    return ea(l), e = rc(
      t,
      l,
      e,
      a,
      void 0,
      u
    ), a = oc(), t !== null && !Jt ? (dc(t, l, u), ce(t, l, u)) : (ht && a && Ji(l), l.flags |= 1, It(t, l, e, u), l.child);
  }
  function $r(t, l, e, a, u, n) {
    return ea(l), l.updateQueue = null, e = ks(
      l,
      a,
      e,
      u
    ), Js(t), a = oc(), t !== null && !Jt ? (dc(t, l, n), ce(t, l, n)) : (ht && a && Ji(l), l.flags |= 1, It(t, l, e, n), l.child);
  }
  function Wr(t, l, e, a, u) {
    if (ea(l), l.stateNode === null) {
      var n = Na, i = e.contextType;
      typeof i == "object" && i !== null && (n = ul(i)), n = new e(a, n), l.memoizedState = n.state !== null && n.state !== void 0 ? n.state : null, n.updater = zc, l.stateNode = n, n._reactInternals = l, n = l.stateNode, n.props = a, n.state = l.memoizedState, n.refs = {}, ac(l), i = e.contextType, n.context = typeof i == "object" && i !== null ? ul(i) : Na, n.state = l.memoizedState, i = e.getDerivedStateFromProps, typeof i == "function" && (Ac(
        l,
        e,
        i,
        a
      ), n.state = l.memoizedState), typeof e.getDerivedStateFromProps == "function" || typeof n.getSnapshotBeforeUpdate == "function" || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (i = n.state, typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount(), i !== n.state && zc.enqueueReplaceState(n, n.state, null), ju(l, a, n, u), xu(), n.state = l.memoizedState), typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !0;
    } else if (t === null) {
      n = l.stateNode;
      var f = l.memoizedProps, s = na(e, f);
      n.props = s;
      var m = n.context, T = e.contextType;
      i = Na, typeof T == "object" && T !== null && (i = ul(T));
      var z = e.getDerivedStateFromProps;
      T = typeof z == "function" || typeof n.getSnapshotBeforeUpdate == "function", f = l.pendingProps !== f, T || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (f || m !== i) && qr(
        l,
        n,
        a,
        i
      ), pe = !1;
      var g = l.memoizedState;
      n.state = g, ju(l, a, n, u), xu(), m = l.memoizedState, f || g !== m || pe ? (typeof z == "function" && (Ac(
        l,
        e,
        z,
        a
      ), m = l.memoizedState), (s = pe || Hr(
        l,
        e,
        s,
        a,
        g,
        m,
        i
      )) ? (T || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount()), typeof n.componentDidMount == "function" && (l.flags |= 4194308)) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), l.memoizedProps = a, l.memoizedState = m), n.props = a, n.state = m, n.context = i, a = s) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !1);
    } else {
      n = l.stateNode, uc(t, l), i = l.memoizedProps, T = na(e, i), n.props = T, z = l.pendingProps, g = n.context, m = e.contextType, s = Na, typeof m == "object" && m !== null && (s = ul(m)), f = e.getDerivedStateFromProps, (m = typeof f == "function" || typeof n.getSnapshotBeforeUpdate == "function") || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (i !== z || g !== s) && qr(
        l,
        n,
        a,
        s
      ), pe = !1, g = l.memoizedState, n.state = g, ju(l, a, n, u), xu();
      var S = l.memoizedState;
      i !== z || g !== S || pe || t !== null && t.dependencies !== null && Sn(t.dependencies) ? (typeof f == "function" && (Ac(
        l,
        e,
        f,
        a
      ), S = l.memoizedState), (T = pe || Hr(
        l,
        e,
        T,
        a,
        g,
        S,
        s
      ) || t !== null && t.dependencies !== null && Sn(t.dependencies)) ? (m || typeof n.UNSAFE_componentWillUpdate != "function" && typeof n.componentWillUpdate != "function" || (typeof n.componentWillUpdate == "function" && n.componentWillUpdate(a, S, s), typeof n.UNSAFE_componentWillUpdate == "function" && n.UNSAFE_componentWillUpdate(
        a,
        S,
        s
      )), typeof n.componentDidUpdate == "function" && (l.flags |= 4), typeof n.getSnapshotBeforeUpdate == "function" && (l.flags |= 1024)) : (typeof n.componentDidUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), l.memoizedProps = a, l.memoizedState = S), n.props = a, n.state = S, n.context = s, a = T) : (typeof n.componentDidUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || i === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), a = !1);
    }
    return n = a, Bn(t, l), a = (l.flags & 128) !== 0, n || a ? (n = l.stateNode, e = a && typeof e.getDerivedStateFromError != "function" ? null : n.render(), l.flags |= 1, t !== null && a ? (l.child = Ya(
      l,
      t.child,
      null,
      u
    ), l.child = Ya(
      l,
      null,
      e,
      u
    )) : It(t, l, e, u), l.memoizedState = n.state, t = l.child) : t = ce(
      t,
      l,
      u
    ), t;
  }
  function Fr(t, l, e, a) {
    return hu(), l.flags |= 256, It(t, l, e, a), l.child;
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
    if ((i = n) || (i = t !== null && t.memoizedState === null ? !1 : (wt.current & 2) !== 0), i && (u = !0, l.flags &= -129), i = (l.flags & 32) !== 0, l.flags &= -33, t === null) {
      if (ht) {
        if (u ? Ee(l) : Ae(), ht) {
          var f = Rt, s;
          if (s = f) {
            t: {
              for (s = f, f = Vl; s.nodeType !== 8; ) {
                if (!f) {
                  f = null;
                  break t;
                }
                if (s = Ql(
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
              treeContext: Fe !== null ? { id: le, overflow: ee } : null,
              retryLane: 536870912,
              hydrationErrors: null
            }, s = jl(
              18,
              null,
              null,
              0
            ), s.stateNode = f, s.return = l, l.child = s, cl = l, Rt = null, s = !0) : s = !1;
          }
          s || ta(l);
        }
        if (f = l.memoizedState, f !== null && (f = f.dehydrated, f !== null))
          return vf(f) ? l.lanes = 32 : l.lanes = 536870912, null;
        ie(l);
      }
      return f = a.children, a = a.fallback, u ? (Ae(), u = l.mode, f = Yn(
        { mode: "hidden", children: f },
        u
      ), a = We(
        a,
        u,
        e,
        null
      ), f.return = l, a.return = l, f.sibling = a, l.child = f, u = l.child, u.memoizedState = Oc(e), u.childLanes = Rc(
        t,
        i,
        e
      ), l.memoizedState = Dc, a) : (Ee(l), _c(l, f));
    }
    if (s = t.memoizedState, s !== null && (f = s.dehydrated, f !== null)) {
      if (n)
        l.flags & 256 ? (Ee(l), l.flags &= -257, l = Uc(
          t,
          l,
          e
        )) : l.memoizedState !== null ? (Ae(), l.child = t.child, l.flags |= 128, l = null) : (Ae(), u = a.fallback, f = l.mode, a = Yn(
          { mode: "visible", children: a.children },
          f
        ), u = We(
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
      else if (Ee(l), vf(f)) {
        if (i = f.nextSibling && f.nextSibling.dataset, i) var m = i.dgst;
        i = m, a = Error(r(419)), a.stack = "", a.digest = i, yu({ value: a, source: null, stack: null }), l = Uc(
          t,
          l,
          e
        );
      } else if (Jt || vu(t, l, e, !1), i = (e & t.childLanes) !== 0, Jt || i) {
        if (i = At, i !== null && (a = e & -e, a = (a & 42) !== 0 ? 1 : Lt(a), a = (a & (i.suspendedLanes | e)) !== 0 ? 0 : a, a !== 0 && a !== s.retryLane))
          throw s.retryLane = a, Ma(t, a), Ml(i, t, a), Lr;
        f.data === "$?" || Fc(), l = Uc(
          t,
          l,
          e
        );
      } else
        f.data === "$?" ? (l.flags |= 192, l.child = t.child, l = null) : (t = s.treeContext, Rt = Ql(
          f.nextSibling
        ), cl = l, ht = !0, Ie = null, Vl = !1, t !== null && (_l[Ul++] = le, _l[Ul++] = ee, _l[Ul++] = Fe, le = t.id, ee = t.overflow, Fe = l), l = _c(
          l,
          a.children
        ), l.flags |= 4096);
      return l;
    }
    return u ? (Ae(), u = a.fallback, f = l.mode, s = t.child, m = s.sibling, a = te(s, {
      mode: "hidden",
      children: a.children
    }), a.subtreeFlags = s.subtreeFlags & 65011712, m !== null ? u = te(m, u) : (u = We(
      u,
      f,
      e,
      null
    ), u.flags |= 2), u.return = l, a.return = l, a.sibling = u, l.child = a, a = u, u = l.child, f = t.child.memoizedState, f === null ? f = Oc(e) : (s = f.cachePool, s !== null ? (m = Vt._currentValue, s = s.parent !== m ? { parent: m, pool: m } : s) : s = Ys(), f = {
      baseLanes: f.baseLanes | e,
      cachePool: s
    }), u.memoizedState = f, u.childLanes = Rc(
      t,
      i,
      e
    ), l.memoizedState = Dc, a) : (Ee(l), e = t.child, t = e.sibling, e = te(e, {
      mode: "visible",
      children: a.children
    }), e.return = l, e.sibling = null, t !== null && (i = l.deletions, i === null ? (l.deletions = [t], l.flags |= 16) : i.push(t)), l.child = e, l.memoizedState = null, e);
  }
  function _c(t, l) {
    return l = Yn(
      { mode: "visible", children: l },
      t.mode
    ), l.return = t, t.child = l;
  }
  function Yn(t, l) {
    return t = jl(22, t, null, l), t.lanes = 0, t.stateNode = {
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
    if (It(t, l, a.children, e), a = wt.current, (a & 2) !== 0)
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
    switch (q(wt, a), u) {
      case "forwards":
        for (e = l.child, u = null; e !== null; )
          t = e.alternate, t !== null && Cn(t) === null && (u = e), e = e.sibling;
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
          if (t = u.alternate, t !== null && Cn(t) === null) {
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
  function ce(t, l, e) {
    if (t !== null && (l.dependencies = t.dependencies), Oe |= l.lanes, (e & l.childLanes) === 0)
      if (t !== null) {
        if (vu(
          t,
          l,
          e,
          !1
        ), (e & l.childLanes) === 0)
          return null;
      } else return null;
    if (t !== null && l.child !== t.child)
      throw Error(r(153));
    if (l.child !== null) {
      for (t = l.child, e = te(t, t.pendingProps), l.child = e, e.return = l; t.sibling !== null; )
        t = t.sibling, e = e.sibling = te(t, t.pendingProps), e.return = l;
      e.sibling = null;
    }
    return l.child;
  }
  function Hc(t, l) {
    return (t.lanes & l) !== 0 ? !0 : (t = t.dependencies, !!(t !== null && Sn(t)));
  }
  function Ih(t, l, e) {
    switch (l.tag) {
      case 3:
        xt(l, l.stateNode.containerInfo), be(l, Vt, t.memoizedState.cache), hu();
        break;
      case 27:
      case 5:
        Pa(l);
        break;
      case 4:
        xt(l, l.stateNode.containerInfo);
        break;
      case 10:
        be(
          l,
          l.type,
          l.memoizedProps.value
        );
        break;
      case 13:
        var a = l.memoizedState;
        if (a !== null)
          return a.dehydrated !== null ? (Ee(l), l.flags |= 128, null) : (e & l.child.childLanes) !== 0 ? Pr(t, l, e) : (Ee(l), t = ce(
            t,
            l,
            e
          ), t !== null ? t.sibling : null);
        Ee(l);
        break;
      case 19:
        var u = (t.flags & 128) !== 0;
        if (a = (e & l.childLanes) !== 0, a || (vu(
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
        if (u = l.memoizedState, u !== null && (u.rendering = null, u.tail = null, u.lastEffect = null), q(wt, wt.current), a) break;
        return null;
      case 22:
      case 23:
        return l.lanes = 0, Jr(t, l, e);
      case 24:
        be(l, Vt, t.memoizedState.cache);
    }
    return ce(t, l, e);
  }
  function lo(t, l, e) {
    if (t !== null)
      if (t.memoizedProps !== l.pendingProps)
        Jt = !0;
      else {
        if (!Hc(t, e) && (l.flags & 128) === 0)
          return Jt = !1, Ih(
            t,
            l,
            e
          );
        Jt = (t.flags & 131072) !== 0;
      }
    else
      Jt = !1, ht && (l.flags & 1048576) !== 0 && Rs(l, gn, l.index);
    switch (l.lanes = 0, l.tag) {
      case 16:
        t: {
          t = l.pendingProps;
          var a = l.elementType, u = a._init;
          if (a = u(a._payload), l.type = a, typeof a == "function")
            Vi(a) ? (t = na(a, t), l.tag = 1, l = Wr(
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
              if (u = a.$$typeof, u === Bt) {
                l.tag = 11, l = Vr(
                  null,
                  l,
                  a,
                  t,
                  e
                );
                break t;
              } else if (u === Ft) {
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
            throw l = $l(a) || a, Error(r(306, l, ""));
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
        return a = l.type, u = na(
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
          if (xt(
            l,
            l.stateNode.containerInfo
          ), t === null) throw Error(r(387));
          a = l.pendingProps;
          var n = l.memoizedState;
          u = n.element, uc(t, l), ju(l, a, null, e);
          var i = l.memoizedState;
          if (a = i.cache, be(l, Vt, a), a !== n.cache && Pi(
            l,
            [Vt],
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
                Error(r(424)),
                l
              ), yu(u), l = Fr(
                t,
                l,
                a,
                e
              );
              break t;
            } else
              for (t = l.stateNode.containerInfo, t.nodeType === 9 ? t = t.body : t = t.nodeName === "HTML" ? t.ownerDocument.body : t, Rt = Ql(t.firstChild), cl = l, ht = !0, Ie = null, Vl = !0, e = Ur(
                l,
                null,
                a,
                e
              ), l.child = e; e; )
                e.flags = e.flags & -3 | 4096, e = e.sibling;
          else {
            if (hu(), a === u) {
              l = ce(
                t,
                l,
                e
              );
              break t;
            }
            It(
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
        return Bn(t, l), t === null ? (e = id(
          l.type,
          null,
          l.pendingProps,
          null
        )) ? l.memoizedState = e : ht || (e = l.type, t = l.pendingProps, a = Pn(
          $.current
        ).createElement(e), a[Yt] = l, a[al] = t, ll(a, e, t), Kt(a), l.stateNode = a) : l.memoizedState = id(
          l.type,
          t.memoizedProps,
          l.pendingProps,
          t.memoizedState
        ), null;
      case 27:
        return Pa(l), t === null && ht && (a = l.stateNode = ad(
          l.type,
          l.pendingProps,
          $.current
        ), cl = l, Vl = !0, u = Rt, Ce(l.type) ? (mf = u, Rt = Ql(
          a.firstChild
        )) : Rt = u), It(
          t,
          l,
          l.pendingProps.children,
          e
        ), Bn(t, l), t === null && (l.flags |= 4194304), l.child;
      case 5:
        return t === null && ht && ((u = a = Rt) && (a = M0(
          a,
          l.type,
          l.pendingProps,
          Vl
        ), a !== null ? (l.stateNode = a, cl = l, Rt = Ql(
          a.firstChild
        ), Vl = !1, u = !0) : u = !1), u || ta(l)), Pa(l), u = l.type, n = l.pendingProps, i = t !== null ? t.memoizedProps : null, a = n.children, df(u, n) ? a = null : i !== null && df(u, i) && (l.flags |= 32), l.memoizedState !== null && (u = rc(
          t,
          l,
          wh,
          null,
          null,
          e
        ), Vu._currentValue = u), Bn(t, l), It(t, l, a, e), l.child;
      case 6:
        return t === null && ht && ((t = e = Rt) && (e = N0(
          e,
          l.pendingProps,
          Vl
        ), e !== null ? (l.stateNode = e, cl = l, Rt = null, t = !0) : t = !1), t || ta(l)), null;
      case 13:
        return Pr(t, l, e);
      case 4:
        return xt(
          l,
          l.stateNode.containerInfo
        ), a = l.pendingProps, t === null ? l.child = Ya(
          l,
          null,
          a,
          e
        ) : It(
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
        return It(
          t,
          l,
          l.pendingProps,
          e
        ), l.child;
      case 8:
        return It(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 12:
        return It(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 10:
        return a = l.pendingProps, be(l, l.type, a.value), It(
          t,
          l,
          a.children,
          e
        ), l.child;
      case 9:
        return u = l.type._context, a = l.pendingProps.children, ea(l), u = ul(u), a = a(u), l.flags |= 1, It(t, l, a, e), l.child;
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
        }, t === null ? (e = Yn(
          a,
          e
        ), e.ref = l.ref, l.child = e, e.return = l, l = e) : (e = te(t.child, a), e.ref = l.ref, l.child = e, e.return = l, l = e), l;
      case 22:
        return Jr(t, l, e);
      case 24:
        return ea(l), a = ul(Vt), t === null ? (u = lc(), u === null && (u = At, n = Ii(), u.pooledCache = n, n.refCount++, n !== null && (u.pooledCacheLanes |= e), u = n), l.memoizedState = {
          parent: a,
          cache: u
        }, ac(l), be(l, Vt, u)) : ((t.lanes & e) !== 0 && (uc(t, l), ju(l, null, null, e), xu()), u = t.memoizedState, n = l.memoizedState, u.parent !== a ? (u = { parent: a, cache: a }, l.memoizedState = u, l.lanes === 0 && (l.memoizedState = l.updateQueue.baseState = u), be(l, Vt, a)) : (a = n.cache, be(l, Vt, a), a !== u.cache && Pi(
          l,
          [Vt],
          e,
          !0
        ))), It(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 29:
        throw l.pendingProps;
    }
    throw Error(r(156, l.tag));
  }
  function fe(t) {
    t.flags |= 4;
  }
  function eo(t, l) {
    if (l.type !== "stylesheet" || (l.state.loading & 4) !== 0)
      t.flags &= -16777217;
    else if (t.flags |= 16777216, !od(l)) {
      if (l = Cl.current, l !== null && ((st & 4194048) === st ? wl !== null : (st & 62914560) !== st && (st & 536870912) === 0 || l !== wl))
        throw bu = ec, Gs;
      t.flags |= 8192;
    }
  }
  function Gn(t, l) {
    l !== null && (t.flags |= 4), t.flags & 16384 && (l = t.tag !== 22 ? me() : 536870912, t.lanes |= l, Za |= l);
  }
  function Du(t, l) {
    if (!ht)
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
  function Mt(t) {
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
        return Mt(l), null;
      case 1:
        return Mt(l), null;
      case 3:
        return e = l.stateNode, a = null, t !== null && (a = t.memoizedState.cache), l.memoizedState.cache !== a && (l.flags |= 2048), ue(Vt), Yl(), e.pendingContext && (e.context = e.pendingContext, e.pendingContext = null), (t === null || t.child === null) && (du(l) ? fe(l) : t === null || t.memoizedState.isDehydrated && (l.flags & 256) === 0 || (l.flags |= 1024, Cs())), Mt(l), null;
      case 26:
        return e = l.memoizedState, t === null ? (fe(l), e !== null ? (Mt(l), eo(l, e)) : (Mt(l), l.flags &= -16777217)) : e ? e !== t.memoizedState ? (fe(l), Mt(l), eo(l, e)) : (Mt(l), l.flags &= -16777217) : (t.memoizedProps !== a && fe(l), Mt(l), l.flags &= -16777217), null;
      case 27:
        ra(l), e = $.current;
        var u = l.type;
        if (t !== null && l.stateNode != null)
          t.memoizedProps !== a && fe(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(r(166));
            return Mt(l), null;
          }
          t = w.current, du(l) ? _s(l) : (t = ad(u, a, e), l.stateNode = t, fe(l));
        }
        return Mt(l), null;
      case 5:
        if (ra(l), e = l.type, t !== null && l.stateNode != null)
          t.memoizedProps !== a && fe(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(r(166));
            return Mt(l), null;
          }
          if (t = w.current, du(l))
            _s(l);
          else {
            switch (u = Pn(
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
            t[Yt] = l, t[al] = a;
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
            t: switch (ll(t, e, a), e) {
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
            t && fe(l);
          }
        }
        return Mt(l), l.flags &= -16777217, null;
      case 6:
        if (t && l.stateNode != null)
          t.memoizedProps !== a && fe(l);
        else {
          if (typeof a != "string" && l.stateNode === null)
            throw Error(r(166));
          if (t = $.current, du(l)) {
            if (t = l.stateNode, e = l.memoizedProps, a = null, u = cl, u !== null)
              switch (u.tag) {
                case 27:
                case 5:
                  a = u.memoizedProps;
              }
            t[Yt] = l, t = !!(t.nodeValue === e || a !== null && a.suppressHydrationWarning === !0 || Wo(t.nodeValue, e)), t || ta(l);
          } else
            t = Pn(t).createTextNode(
              a
            ), t[Yt] = l, l.stateNode = t;
        }
        return Mt(l), null;
      case 13:
        if (a = l.memoizedState, t === null || t.memoizedState !== null && t.memoizedState.dehydrated !== null) {
          if (u = du(l), a !== null && a.dehydrated !== null) {
            if (t === null) {
              if (!u) throw Error(r(318));
              if (u = l.memoizedState, u = u !== null ? u.dehydrated : null, !u) throw Error(r(317));
              u[Yt] = l;
            } else
              hu(), (l.flags & 128) === 0 && (l.memoizedState = null), l.flags |= 4;
            Mt(l), u = !1;
          } else
            u = Cs(), t !== null && t.memoizedState !== null && (t.memoizedState.hydrationErrors = u), u = !0;
          if (!u)
            return l.flags & 256 ? (ie(l), l) : (ie(l), null);
        }
        if (ie(l), (l.flags & 128) !== 0)
          return l.lanes = e, l;
        if (e = a !== null, t = t !== null && t.memoizedState !== null, e) {
          a = l.child, u = null, a.alternate !== null && a.alternate.memoizedState !== null && a.alternate.memoizedState.cachePool !== null && (u = a.alternate.memoizedState.cachePool.pool);
          var n = null;
          a.memoizedState !== null && a.memoizedState.cachePool !== null && (n = a.memoizedState.cachePool.pool), n !== u && (a.flags |= 2048);
        }
        return e !== t && e && (l.child.flags |= 8192), Gn(l, l.updateQueue), Mt(l), null;
      case 4:
        return Yl(), t === null && cf(l.stateNode.containerInfo), Mt(l), null;
      case 10:
        return ue(l.type), Mt(l), null;
      case 19:
        if (Y(wt), u = l.memoizedState, u === null) return Mt(l), null;
        if (a = (l.flags & 128) !== 0, n = u.rendering, n === null)
          if (a) Du(u, !1);
          else {
            if (_t !== 0 || t !== null && (t.flags & 128) !== 0)
              for (t = l.child; t !== null; ) {
                if (n = Cn(t), n !== null) {
                  for (l.flags |= 128, Du(u, !1), t = n.updateQueue, l.updateQueue = t, Gn(l, t), l.subtreeFlags = 0, t = e, e = l.child; e !== null; )
                    Os(e, t), e = e.sibling;
                  return q(
                    wt,
                    wt.current & 1 | 2
                  ), l.child;
                }
                t = t.sibling;
              }
            u.tail !== null && bl() > Zn && (l.flags |= 128, a = !0, Du(u, !1), l.lanes = 4194304);
          }
        else {
          if (!a)
            if (t = Cn(n), t !== null) {
              if (l.flags |= 128, a = !0, t = t.updateQueue, l.updateQueue = t, Gn(l, t), Du(u, !0), u.tail === null && u.tailMode === "hidden" && !n.alternate && !ht)
                return Mt(l), null;
            } else
              2 * bl() - u.renderingStartTime > Zn && e !== 536870912 && (l.flags |= 128, a = !0, Du(u, !1), l.lanes = 4194304);
          u.isBackwards ? (n.sibling = l.child, l.child = n) : (t = u.last, t !== null ? t.sibling = n : l.child = n, u.last = n);
        }
        return u.tail !== null ? (l = u.tail, u.rendering = l, u.tail = l.sibling, u.renderingStartTime = bl(), l.sibling = null, t = wt.current, q(wt, a ? t & 1 | 2 : t & 1), l) : (Mt(l), null);
      case 22:
      case 23:
        return ie(l), fc(), a = l.memoizedState !== null, t !== null ? t.memoizedState !== null !== a && (l.flags |= 8192) : a && (l.flags |= 8192), a ? (e & 536870912) !== 0 && (l.flags & 128) === 0 && (Mt(l), l.subtreeFlags & 6 && (l.flags |= 8192)) : Mt(l), e = l.updateQueue, e !== null && Gn(l, e.retryQueue), e = null, t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), a = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (a = l.memoizedState.cachePool.pool), a !== e && (l.flags |= 2048), t !== null && Y(aa), null;
      case 24:
        return e = null, t !== null && (e = t.memoizedState.cache), l.memoizedState.cache !== e && (l.flags |= 2048), ue(Vt), Mt(l), null;
      case 25:
        return null;
      case 30:
        return null;
    }
    throw Error(r(156, l.tag));
  }
  function l0(t, l) {
    switch (ki(l), l.tag) {
      case 1:
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 3:
        return ue(Vt), Yl(), t = l.flags, (t & 65536) !== 0 && (t & 128) === 0 ? (l.flags = t & -65537 | 128, l) : null;
      case 26:
      case 27:
      case 5:
        return ra(l), null;
      case 13:
        if (ie(l), t = l.memoizedState, t !== null && t.dehydrated !== null) {
          if (l.alternate === null)
            throw Error(r(340));
          hu();
        }
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 19:
        return Y(wt), null;
      case 4:
        return Yl(), null;
      case 10:
        return ue(l.type), null;
      case 22:
      case 23:
        return ie(l), fc(), t !== null && Y(aa), t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 24:
        return ue(Vt), null;
      case 25:
        return null;
      default:
        return null;
    }
  }
  function ao(t, l) {
    switch (ki(l), l.tag) {
      case 3:
        ue(Vt), Yl();
        break;
      case 26:
      case 27:
      case 5:
        ra(l);
        break;
      case 4:
        Yl();
        break;
      case 13:
        ie(l);
        break;
      case 19:
        Y(wt);
        break;
      case 10:
        ue(l.type);
        break;
      case 22:
      case 23:
        ie(l), fc(), t !== null && Y(aa);
        break;
      case 24:
        ue(Vt);
    }
  }
  function Ou(t, l) {
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
      jt(l, l.return, f);
    }
  }
  function ze(t, l, e) {
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
              } catch (T) {
                jt(
                  u,
                  s,
                  T
                );
              }
            }
          }
          a = a.next;
        } while (a !== n);
      }
    } catch (T) {
      jt(l, l.return, T);
    }
  }
  function uo(t) {
    var l = t.updateQueue;
    if (l !== null) {
      var e = t.stateNode;
      try {
        ws(l, e);
      } catch (a) {
        jt(t, t.return, a);
      }
    }
  }
  function no(t, l, e) {
    e.props = na(
      t.type,
      t.memoizedProps
    ), e.state = t.memoizedState;
    try {
      e.componentWillUnmount();
    } catch (a) {
      jt(t, l, a);
    }
  }
  function Ru(t, l) {
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
      jt(t, l, u);
    }
  }
  function Kl(t, l) {
    var e = t.ref, a = t.refCleanup;
    if (e !== null)
      if (typeof a == "function")
        try {
          a();
        } catch (u) {
          jt(t, l, u);
        } finally {
          t.refCleanup = null, t = t.alternate, t != null && (t.refCleanup = null);
        }
      else if (typeof e == "function")
        try {
          e(null);
        } catch (u) {
          jt(t, l, u);
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
      jt(t, t.return, u);
    }
  }
  function qc(t, l, e) {
    try {
      var a = t.stateNode;
      j0(a, t.type, e, l), a[al] = l;
    } catch (u) {
      jt(t, t.return, u);
    }
  }
  function co(t) {
    return t.tag === 5 || t.tag === 3 || t.tag === 26 || t.tag === 27 && Ce(t.type) || t.tag === 4;
  }
  function Bc(t) {
    t: for (; ; ) {
      for (; t.sibling === null; ) {
        if (t.return === null || co(t.return)) return null;
        t = t.return;
      }
      for (t.sibling.return = t.return, t = t.sibling; t.tag !== 5 && t.tag !== 6 && t.tag !== 18; ) {
        if (t.tag === 27 && Ce(t.type) || t.flags & 2 || t.child === null || t.tag === 4) continue t;
        t.child.return = t, t = t.child;
      }
      if (!(t.flags & 2)) return t.stateNode;
    }
  }
  function Yc(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? (e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e).insertBefore(t, l) : (l = e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, l.appendChild(t), e = e._reactRootContainer, e != null || l.onclick !== null || (l.onclick = Fn));
    else if (a !== 4 && (a === 27 && Ce(t.type) && (e = t.stateNode, l = null), t = t.child, t !== null))
      for (Yc(t, l, e), t = t.sibling; t !== null; )
        Yc(t, l, e), t = t.sibling;
  }
  function Xn(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? e.insertBefore(t, l) : e.appendChild(t);
    else if (a !== 4 && (a === 27 && Ce(t.type) && (e = t.stateNode), t = t.child, t !== null))
      for (Xn(t, l, e), t = t.sibling; t !== null; )
        Xn(t, l, e), t = t.sibling;
  }
  function fo(t) {
    var l = t.stateNode, e = t.memoizedProps;
    try {
      for (var a = t.type, u = l.attributes; u.length; )
        l.removeAttributeNode(u[0]);
      ll(l, a, e), l[Yt] = t, l[al] = e;
    } catch (n) {
      jt(t, t.return, n);
    }
  }
  var se = !1, Ht = !1, Gc = !1, so = typeof WeakSet == "function" ? WeakSet : Set, kt = null;
  function e0(t, l) {
    if (t = t.containerInfo, rf = ui, t = ps(t), Bi(t)) {
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
            var i = 0, f = -1, s = -1, m = 0, T = 0, z = t, g = null;
            l: for (; ; ) {
              for (var S; z !== e || u !== 0 && z.nodeType !== 3 || (f = i + u), z !== n || a !== 0 && z.nodeType !== 3 || (s = i + a), z.nodeType === 3 && (i += z.nodeValue.length), (S = z.firstChild) !== null; )
                g = z, z = S;
              for (; ; ) {
                if (z === t) break l;
                if (g === e && ++m === u && (f = i), g === n && ++T === a && (s = i), (S = z.nextSibling) !== null) break;
                z = g, g = z.parentNode;
              }
              z = S;
            }
            e = f === -1 || s === -1 ? null : { start: f, end: s };
          } else e = null;
        }
      e = e || { start: 0, end: 0 };
    } else e = null;
    for (of = { focusedElem: t, selectionRange: e }, ui = !1, kt = l; kt !== null; )
      if (l = kt, t = l.child, (l.subtreeFlags & 1024) !== 0 && t !== null)
        t.return = l, kt = t;
      else
        for (; kt !== null; ) {
          switch (l = kt, n = l.alternate, t = l.flags, l.tag) {
            case 0:
              break;
            case 11:
            case 15:
              break;
            case 1:
              if ((t & 1024) !== 0 && n !== null) {
                t = void 0, e = l, u = n.memoizedProps, n = n.memoizedState, a = e.stateNode;
                try {
                  var W = na(
                    e.type,
                    u,
                    e.elementType === e.type
                  );
                  t = a.getSnapshotBeforeUpdate(
                    W,
                    n
                  ), a.__reactInternalSnapshotBeforeUpdate = t;
                } catch (J) {
                  jt(
                    e,
                    e.return,
                    J
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
              if ((t & 1024) !== 0) throw Error(r(163));
          }
          if (t = l.sibling, t !== null) {
            t.return = l.return, kt = t;
            break;
          }
          kt = l.return;
        }
  }
  function ro(t, l, e) {
    var a = e.flags;
    switch (e.tag) {
      case 0:
      case 11:
      case 15:
        Me(t, e), a & 4 && Ou(5, e);
        break;
      case 1:
        if (Me(t, e), a & 4)
          if (t = e.stateNode, l === null)
            try {
              t.componentDidMount();
            } catch (i) {
              jt(e, e.return, i);
            }
          else {
            var u = na(
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
              jt(
                e,
                e.return,
                i
              );
            }
          }
        a & 64 && uo(e), a & 512 && Ru(e, e.return);
        break;
      case 3:
        if (Me(t, e), a & 64 && (t = e.updateQueue, t !== null)) {
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
            jt(e, e.return, i);
          }
        }
        break;
      case 27:
        l === null && a & 4 && fo(e);
      case 26:
      case 5:
        Me(t, e), l === null && a & 4 && io(e), a & 512 && Ru(e, e.return);
        break;
      case 12:
        Me(t, e);
        break;
      case 13:
        Me(t, e), a & 4 && yo(t, e), a & 64 && (t = e.memoizedState, t !== null && (t = t.dehydrated, t !== null && (e = o0.bind(
          null,
          e
        ), D0(t, e))));
        break;
      case 22:
        if (a = e.memoizedState !== null || se, !a) {
          l = l !== null && l.memoizedState !== null || Ht, u = se;
          var n = Ht;
          se = a, (Ht = l) && !n ? Ne(
            t,
            e,
            (e.subtreeFlags & 8772) !== 0
          ) : Me(t, e), se = u, Ht = n;
        }
        break;
      case 30:
        break;
      default:
        Me(t, e);
    }
  }
  function oo(t) {
    var l = t.alternate;
    l !== null && (t.alternate = null, oo(l)), t.child = null, t.deletions = null, t.sibling = null, t.tag === 5 && (l = t.stateNode, l !== null && Si(l)), t.stateNode = null, t.return = null, t.dependencies = null, t.memoizedProps = null, t.memoizedState = null, t.pendingProps = null, t.stateNode = null, t.updateQueue = null;
  }
  var zt = null, vl = !1;
  function re(t, l, e) {
    for (e = e.child; e !== null; )
      ho(t, l, e), e = e.sibling;
  }
  function ho(t, l, e) {
    if (ft && typeof ft.onCommitFiberUnmount == "function")
      try {
        ft.onCommitFiberUnmount(ol, e);
      } catch {
      }
    switch (e.tag) {
      case 26:
        Ht || Kl(e, l), re(
          t,
          l,
          e
        ), e.memoizedState ? e.memoizedState.count-- : e.stateNode && (e = e.stateNode, e.parentNode.removeChild(e));
        break;
      case 27:
        Ht || Kl(e, l);
        var a = zt, u = vl;
        Ce(e.type) && (zt = e.stateNode, vl = !1), re(
          t,
          l,
          e
        ), Xu(e.stateNode), zt = a, vl = u;
        break;
      case 5:
        Ht || Kl(e, l);
      case 6:
        if (a = zt, u = vl, zt = null, re(
          t,
          l,
          e
        ), zt = a, vl = u, zt !== null)
          if (vl)
            try {
              (zt.nodeType === 9 ? zt.body : zt.nodeName === "HTML" ? zt.ownerDocument.body : zt).removeChild(e.stateNode);
            } catch (n) {
              jt(
                e,
                l,
                n
              );
            }
          else
            try {
              zt.removeChild(e.stateNode);
            } catch (n) {
              jt(
                e,
                l,
                n
              );
            }
        break;
      case 18:
        zt !== null && (vl ? (t = zt, ld(
          t.nodeType === 9 ? t.body : t.nodeName === "HTML" ? t.ownerDocument.body : t,
          e.stateNode
        ), ku(t)) : ld(zt, e.stateNode));
        break;
      case 4:
        a = zt, u = vl, zt = e.stateNode.containerInfo, vl = !0, re(
          t,
          l,
          e
        ), zt = a, vl = u;
        break;
      case 0:
      case 11:
      case 14:
      case 15:
        Ht || ze(2, e, l), Ht || ze(4, e, l), re(
          t,
          l,
          e
        );
        break;
      case 1:
        Ht || (Kl(e, l), a = e.stateNode, typeof a.componentWillUnmount == "function" && no(
          e,
          l,
          a
        )), re(
          t,
          l,
          e
        );
        break;
      case 21:
        re(
          t,
          l,
          e
        );
        break;
      case 22:
        Ht = (a = Ht) || e.memoizedState !== null, re(
          t,
          l,
          e
        ), Ht = a;
        break;
      default:
        re(
          t,
          l,
          e
        );
    }
  }
  function yo(t, l) {
    if (l.memoizedState === null && (t = l.alternate, t !== null && (t = t.memoizedState, t !== null && (t = t.dehydrated, t !== null))))
      try {
        ku(t);
      } catch (e) {
        jt(l, l.return, e);
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
        throw Error(r(435, t.tag));
    }
  }
  function Xc(t, l) {
    var e = a0(t);
    l.forEach(function(a) {
      var u = d0.bind(null, t, a);
      e.has(a) || (e.add(a), a.then(u, u));
    });
  }
  function Tl(t, l) {
    var e = l.deletions;
    if (e !== null)
      for (var a = 0; a < e.length; a++) {
        var u = e[a], n = t, i = l, f = i;
        t: for (; f !== null; ) {
          switch (f.tag) {
            case 27:
              if (Ce(f.type)) {
                zt = f.stateNode, vl = !1;
                break t;
              }
              break;
            case 5:
              zt = f.stateNode, vl = !1;
              break t;
            case 3:
            case 4:
              zt = f.stateNode.containerInfo, vl = !0;
              break t;
          }
          f = f.return;
        }
        if (zt === null) throw Error(r(160));
        ho(n, i, u), zt = null, vl = !1, n = u.alternate, n !== null && (n.return = null), u.return = null;
      }
    if (l.subtreeFlags & 13878)
      for (l = l.child; l !== null; )
        vo(l, t), l = l.sibling;
  }
  var Xl = null;
  function vo(t, l) {
    var e = t.alternate, a = t.flags;
    switch (t.tag) {
      case 0:
      case 11:
      case 14:
      case 15:
        Tl(l, t), El(t), a & 4 && (ze(3, t, t.return), Ou(3, t), ze(5, t, t.return));
        break;
      case 1:
        Tl(l, t), El(t), a & 512 && (Ht || e === null || Kl(e, e.return)), a & 64 && se && (t = t.updateQueue, t !== null && (a = t.callbacks, a !== null && (e = t.shared.hiddenCallbacks, t.shared.hiddenCallbacks = e === null ? a : e.concat(a))));
        break;
      case 26:
        var u = Xl;
        if (Tl(l, t), El(t), a & 512 && (Ht || e === null || Kl(e, e.return)), a & 4) {
          var n = e !== null ? e.memoizedState : null;
          if (a = t.memoizedState, e === null)
            if (a === null)
              if (t.stateNode === null) {
                t: {
                  a = t.type, e = t.memoizedProps, u = u.ownerDocument || u;
                  l: switch (a) {
                    case "title":
                      n = u.getElementsByTagName("title")[0], (!n || n[lu] || n[Yt] || n.namespaceURI === "http://www.w3.org/2000/svg" || n.hasAttribute("itemprop")) && (n = u.createElement(a), u.head.insertBefore(
                        n,
                        u.querySelector("head > title")
                      )), ll(n, a, e), n[Yt] = t, Kt(n), a = n;
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
                      n = u.createElement(a), ll(n, a, e), u.head.appendChild(n);
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
                      n = u.createElement(a), ll(n, a, e), u.head.appendChild(n);
                      break;
                    default:
                      throw Error(r(468, a));
                  }
                  n[Yt] = t, Kt(n), a = n;
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
        Tl(l, t), El(t), a & 512 && (Ht || e === null || Kl(e, e.return)), e !== null && a & 4 && qc(
          t,
          t.memoizedProps,
          e.memoizedProps
        );
        break;
      case 5:
        if (Tl(l, t), El(t), a & 512 && (Ht || e === null || Kl(e, e.return)), t.flags & 32) {
          u = t.stateNode;
          try {
            pa(u, "");
          } catch (S) {
            jt(t, t.return, S);
          }
        }
        a & 4 && t.stateNode != null && (u = t.memoizedProps, qc(
          t,
          u,
          e !== null ? e.memoizedProps : u
        )), a & 1024 && (Gc = !0);
        break;
      case 6:
        if (Tl(l, t), El(t), a & 4) {
          if (t.stateNode === null)
            throw Error(r(162));
          a = t.memoizedProps, e = t.stateNode;
          try {
            e.nodeValue = a;
          } catch (S) {
            jt(t, t.return, S);
          }
        }
        break;
      case 3:
        if (li = null, u = Xl, Xl = In(l.containerInfo), Tl(l, t), Xl = u, El(t), a & 4 && e !== null && e.memoizedState.isDehydrated)
          try {
            ku(l.containerInfo);
          } catch (S) {
            jt(t, t.return, S);
          }
        Gc && (Gc = !1, mo(t));
        break;
      case 4:
        a = Xl, Xl = In(
          t.stateNode.containerInfo
        ), Tl(l, t), El(t), Xl = a;
        break;
      case 12:
        Tl(l, t), El(t);
        break;
      case 13:
        Tl(l, t), El(t), t.child.flags & 8192 && t.memoizedState !== null != (e !== null && e.memoizedState !== null) && (Kc = bl()), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Xc(t, a)));
        break;
      case 22:
        u = t.memoizedState !== null;
        var s = e !== null && e.memoizedState !== null, m = se, T = Ht;
        if (se = m || u, Ht = T || s, Tl(l, t), Ht = T, se = m, El(t), a & 8192)
          t: for (l = t.stateNode, l._visibility = u ? l._visibility & -2 : l._visibility | 1, u && (e === null || s || se || Ht || ia(t)), e = null, l = t; ; ) {
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
                  jt(s, s.return, S);
                }
              }
            } else if (l.tag === 6) {
              if (e === null) {
                s = l;
                try {
                  s.stateNode.nodeValue = u ? "" : s.memoizedProps;
                } catch (S) {
                  jt(s, s.return, S);
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
        Tl(l, t), El(t), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Xc(t, a)));
        break;
      case 30:
        break;
      case 21:
        break;
      default:
        Tl(l, t), El(t);
    }
  }
  function El(t) {
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
        if (e == null) throw Error(r(160));
        switch (e.tag) {
          case 27:
            var u = e.stateNode, n = Bc(t);
            Xn(t, n, u);
            break;
          case 5:
            var i = e.stateNode;
            e.flags & 32 && (pa(i, ""), e.flags &= -33);
            var f = Bc(t);
            Xn(t, f, i);
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
            throw Error(r(161));
        }
      } catch (T) {
        jt(t, t.return, T);
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
  function Me(t, l) {
    if (l.subtreeFlags & 8772)
      for (l = l.child; l !== null; )
        ro(t, l.alternate, l), l = l.sibling;
  }
  function ia(t) {
    for (t = t.child; t !== null; ) {
      var l = t;
      switch (l.tag) {
        case 0:
        case 11:
        case 14:
        case 15:
          ze(4, l, l.return), ia(l);
          break;
        case 1:
          Kl(l, l.return);
          var e = l.stateNode;
          typeof e.componentWillUnmount == "function" && no(
            l,
            l.return,
            e
          ), ia(l);
          break;
        case 27:
          Xu(l.stateNode);
        case 26:
        case 5:
          Kl(l, l.return), ia(l);
          break;
        case 22:
          l.memoizedState === null && ia(l);
          break;
        case 30:
          ia(l);
          break;
        default:
          ia(l);
      }
      t = t.sibling;
    }
  }
  function Ne(t, l, e) {
    for (e = e && (l.subtreeFlags & 8772) !== 0, l = l.child; l !== null; ) {
      var a = l.alternate, u = t, n = l, i = n.flags;
      switch (n.tag) {
        case 0:
        case 11:
        case 15:
          Ne(
            u,
            n,
            e
          ), Ou(4, n);
          break;
        case 1:
          if (Ne(
            u,
            n,
            e
          ), a = n, u = a.stateNode, typeof u.componentDidMount == "function")
            try {
              u.componentDidMount();
            } catch (m) {
              jt(a, a.return, m);
            }
          if (a = n, u = a.updateQueue, u !== null) {
            var f = a.stateNode;
            try {
              var s = u.shared.hiddenCallbacks;
              if (s !== null)
                for (u.shared.hiddenCallbacks = null, u = 0; u < s.length; u++)
                  Vs(s[u], f);
            } catch (m) {
              jt(a, a.return, m);
            }
          }
          e && i & 64 && uo(n), Ru(n, n.return);
          break;
        case 27:
          fo(n);
        case 26:
        case 5:
          Ne(
            u,
            n,
            e
          ), e && a === null && i & 4 && io(n), Ru(n, n.return);
          break;
        case 12:
          Ne(
            u,
            n,
            e
          );
          break;
        case 13:
          Ne(
            u,
            n,
            e
          ), e && i & 4 && yo(u, n);
          break;
        case 22:
          n.memoizedState === null && Ne(
            u,
            n,
            e
          ), Ru(n, n.return);
          break;
        case 30:
          break;
        default:
          Ne(
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
    t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), t = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (t = l.memoizedState.cachePool.pool), t !== e && (t != null && t.refCount++, e != null && mu(e));
  }
  function Zc(t, l) {
    t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && mu(t));
  }
  function Jl(t, l, e, a) {
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
        Jl(
          t,
          l,
          e,
          a
        ), u & 2048 && Ou(9, l);
        break;
      case 1:
        Jl(
          t,
          l,
          e,
          a
        );
        break;
      case 3:
        Jl(
          t,
          l,
          e,
          a
        ), u & 2048 && (t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && mu(t)));
        break;
      case 12:
        if (u & 2048) {
          Jl(
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
            jt(l, l.return, s);
          }
        } else
          Jl(
            t,
            l,
            e,
            a
          );
        break;
      case 13:
        Jl(
          t,
          l,
          e,
          a
        );
        break;
      case 23:
        break;
      case 22:
        n = l.stateNode, i = l.alternate, l.memoizedState !== null ? n._visibility & 2 ? Jl(
          t,
          l,
          e,
          a
        ) : _u(t, l) : n._visibility & 2 ? Jl(
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
        Jl(
          t,
          l,
          e,
          a
        ), u & 2048 && Zc(l.alternate, l);
        break;
      default:
        Jl(
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
          ), Ou(8, i);
          break;
        case 23:
          break;
        case 22:
          var T = i.stateNode;
          i.memoizedState !== null ? T._visibility & 2 ? Ga(
            n,
            i,
            f,
            s,
            u
          ) : _u(
            n,
            i
          ) : (T._visibility |= 2, Ga(
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
  function _u(t, l) {
    if (l.subtreeFlags & 10256)
      for (l = l.child; l !== null; ) {
        var e = t, a = l, u = a.flags;
        switch (a.tag) {
          case 22:
            _u(e, a), u & 2048 && Qc(
              a.alternate,
              a
            );
            break;
          case 24:
            _u(e, a), u & 2048 && Zc(a.alternate, a);
            break;
          default:
            _u(e, a);
        }
        l = l.sibling;
      }
  }
  var Uu = 8192;
  function Xa(t) {
    if (t.subtreeFlags & Uu)
      for (t = t.child; t !== null; )
        So(t), t = t.sibling;
  }
  function So(t) {
    switch (t.tag) {
      case 26:
        Xa(t), t.flags & Uu && t.memoizedState !== null && Z0(
          Xl,
          t.memoizedState,
          t.memoizedProps
        );
        break;
      case 5:
        Xa(t);
        break;
      case 3:
      case 4:
        var l = Xl;
        Xl = In(t.stateNode.containerInfo), Xa(t), Xl = l;
        break;
      case 22:
        t.memoizedState === null && (l = t.alternate, l !== null && l.memoizedState !== null ? (l = Uu, Uu = 16777216, Xa(t), Uu = l) : Xa(t));
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
  function Cu(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          kt = a, xo(
            a,
            t
          );
        }
      bo(t);
    }
    if (t.subtreeFlags & 10256)
      for (t = t.child; t !== null; )
        po(t), t = t.sibling;
  }
  function po(t) {
    switch (t.tag) {
      case 0:
      case 11:
      case 15:
        Cu(t), t.flags & 2048 && ze(9, t, t.return);
        break;
      case 3:
        Cu(t);
        break;
      case 12:
        Cu(t);
        break;
      case 22:
        var l = t.stateNode;
        t.memoizedState !== null && l._visibility & 2 && (t.return === null || t.return.tag !== 13) ? (l._visibility &= -3, Qn(t)) : Cu(t);
        break;
      default:
        Cu(t);
    }
  }
  function Qn(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          kt = a, xo(
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
          ze(8, l, l.return), Qn(l);
          break;
        case 22:
          e = l.stateNode, e._visibility & 2 && (e._visibility &= -3, Qn(l));
          break;
        default:
          Qn(l);
      }
      t = t.sibling;
    }
  }
  function xo(t, l) {
    for (; kt !== null; ) {
      var e = kt;
      switch (e.tag) {
        case 0:
        case 11:
        case 15:
          ze(8, e, l);
          break;
        case 23:
        case 22:
          if (e.memoizedState !== null && e.memoizedState.cachePool !== null) {
            var a = e.memoizedState.cachePool.pool;
            a != null && a.refCount++;
          }
          break;
        case 24:
          mu(e.memoizedState.cache);
      }
      if (a = e.child, a !== null) a.return = e, kt = a;
      else
        t: for (e = t; kt !== null; ) {
          a = kt;
          var u = a.sibling, n = a.return;
          if (oo(a), a === e) {
            kt = null;
            break t;
          }
          if (u !== null) {
            u.return = n, kt = u;
            break t;
          }
          kt = n;
        }
    }
  }
  var u0 = {
    getCacheForType: function(t) {
      var l = ul(Vt), e = l.data.get(t);
      return e === void 0 && (e = t(), l.data.set(t, e)), e;
    }
  }, n0 = typeof WeakMap == "function" ? WeakMap : Map, vt = 0, At = null, ut = null, st = 0, mt = 0, Al = null, De = !1, Qa = !1, Lc = !1, oe = 0, _t = 0, Oe = 0, ca = 0, Vc = 0, Hl = 0, Za = 0, Hu = null, ml = null, wc = !1, Kc = 0, Zn = 1 / 0, Ln = null, Re = null, tl = 0, _e = null, La = null, Va = 0, Jc = 0, kc = null, jo = null, qu = 0, $c = null;
  function zl() {
    if ((vt & 2) !== 0 && st !== 0)
      return st & -st;
    if (x.T !== null) {
      var t = Ra;
      return t !== 0 ? t : ef();
    }
    return dl();
  }
  function To() {
    Hl === 0 && (Hl = (st & 536870912) === 0 || ht ? we() : 536870912);
    var t = Cl.current;
    return t !== null && (t.flags |= 32), Hl;
  }
  function Ml(t, l, e) {
    (t === At && (mt === 2 || mt === 9) || t.cancelPendingCommit !== null) && (wa(t, 0), Ue(
      t,
      st,
      Hl,
      !1
    )), X(t, e), ((vt & 2) === 0 || t !== At) && (t === At && ((vt & 2) === 0 && (ca |= e), _t === 4 && Ue(
      t,
      st,
      Hl,
      !1
    )), kl(t));
  }
  function Eo(t, l, e) {
    if ((vt & 6) !== 0) throw Error(r(327));
    var a = !e && (l & 124) === 0 && (l & t.expiredLanes) === 0 || Ve(t, l), u = a ? f0(t, l) : Pc(t, l, !0), n = a;
    do {
      if (u === 0) {
        Qa && !a && Ue(t, l, 0, !1);
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
              u = Hu;
              var s = f.current.memoizedState.isDehydrated;
              if (s && (wa(f, i).flags |= 256), i = Pc(
                f,
                i,
                !1
              ), i !== 2) {
                if (Lc && !s) {
                  f.errorRecoveryDisabledLanes |= n, ca |= n, u = 4;
                  break t;
                }
                n = ml, ml = u, n !== null && (ml === null ? ml = n : ml.push.apply(
                  ml,
                  n
                ));
              }
              u = i;
            }
            if (n = !1, u !== 2) continue;
          }
        }
        if (u === 1) {
          wa(t, 0), Ue(t, l, 0, !0);
          break;
        }
        t: {
          switch (a = t, n = u, n) {
            case 0:
            case 1:
              throw Error(r(345));
            case 4:
              if ((l & 4194048) !== l) break;
            case 6:
              Ue(
                a,
                l,
                Hl,
                !De
              );
              break t;
            case 2:
              ml = null;
              break;
            case 3:
            case 5:
              break;
            default:
              throw Error(r(329));
          }
          if ((l & 62914560) === l && (u = Kc + 300 - bl(), 10 < u)) {
            if (Ue(
              a,
              l,
              Hl,
              !De
            ), Le(a, 0, !0) !== 0) break t;
            a.timeoutHandle = Io(
              Ao.bind(
                null,
                a,
                e,
                ml,
                Ln,
                wc,
                l,
                Hl,
                ca,
                Za,
                De,
                n,
                2,
                -0,
                0
              ),
              u
            );
            break t;
          }
          Ao(
            a,
            e,
            ml,
            Ln,
            wc,
            l,
            Hl,
            ca,
            Za,
            De,
            n,
            0,
            -0,
            0
          );
        }
      }
      break;
    } while (!0);
    kl(t);
  }
  function Ao(t, l, e, a, u, n, i, f, s, m, T, z, g, S) {
    if (t.timeoutHandle = -1, z = l.subtreeFlags, (z & 8192 || (z & 16785408) === 16785408) && (Lu = { stylesheets: null, count: 0, unsuspend: Q0 }, So(l), z = L0(), z !== null)) {
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
          T,
          1,
          g,
          S
        )
      ), Ue(t, n, i, !m);
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
  function Ue(t, l, e, a) {
    l &= ~Vc, l &= ~ca, t.suspendedLanes |= l, t.pingedLanes &= ~l, a && (t.warmLanes |= l), a = t.expirationTimes;
    for (var u = l; 0 < u; ) {
      var n = 31 - Pt(u), i = 1 << n;
      a[n] = -1, u &= ~i;
    }
    e !== 0 && it(t, e, l);
  }
  function Vn() {
    return (vt & 6) === 0 ? (Bu(0), !1) : !0;
  }
  function Wc() {
    if (ut !== null) {
      if (mt === 0)
        var t = ut.return;
      else
        t = ut, ae = la = null, hc(t), Ba = null, Mu = 0, t = ut;
      for (; t !== null; )
        ao(t.alternate, t), t = t.return;
      ut = null;
    }
  }
  function wa(t, l) {
    var e = t.timeoutHandle;
    e !== -1 && (t.timeoutHandle = -1, E0(e)), e = t.cancelPendingCommit, e !== null && (t.cancelPendingCommit = null, e()), Wc(), At = t, ut = e = te(t.current, null), st = l, mt = 0, Al = null, De = !1, Qa = Ve(t, l), Lc = !1, Za = Hl = Vc = ca = Oe = _t = 0, ml = Hu = null, wc = !1, (l & 8) !== 0 && (l |= l & 32);
    var a = t.entangledLanes;
    if (a !== 0)
      for (t = t.entanglements, a &= l; 0 < a; ) {
        var u = 31 - Pt(a), n = 1 << u;
        l |= t[u], a &= ~n;
      }
    return oe = l, dn(), e;
  }
  function zo(t, l) {
    lt = null, x.H = Rn, l === Su || l === xn ? (l = Zs(), mt = 3) : l === Gs ? (l = Zs(), mt = 4) : mt = l === Lr ? 8 : l !== null && typeof l == "object" && typeof l.then == "function" ? 6 : 1, Al = l, ut === null && (_t = 1, qn(
      t,
      Ol(l, t.current)
    ));
  }
  function Mo() {
    var t = x.H;
    return x.H = Rn, t === null ? Rn : t;
  }
  function No() {
    var t = x.A;
    return x.A = u0, t;
  }
  function Fc() {
    _t = 4, De || (st & 4194048) !== st && Cl.current !== null || (Qa = !0), (Oe & 134217727) === 0 && (ca & 134217727) === 0 || At === null || Ue(
      At,
      st,
      Hl,
      !1
    );
  }
  function Pc(t, l, e) {
    var a = vt;
    vt |= 2;
    var u = Mo(), n = No();
    (At !== t || st !== l) && (Ln = null, wa(t, l)), l = !1;
    var i = _t;
    t: do
      try {
        if (mt !== 0 && ut !== null) {
          var f = ut, s = Al;
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
        c0(), i = _t;
        break;
      } catch (T) {
        zo(t, T);
      }
    while (!0);
    return l && t.shellSuspendCounter++, ae = la = null, vt = a, x.H = u, x.A = n, ut === null && (At = null, st = 0, dn()), i;
  }
  function c0() {
    for (; ut !== null; ) Do(ut);
  }
  function f0(t, l) {
    var e = vt;
    vt |= 2;
    var a = Mo(), u = No();
    At !== t || st !== l ? (Ln = null, Zn = bl() + 500, wa(t, l)) : Qa = Ve(
      t,
      l
    );
    t: do
      try {
        if (mt !== 0 && ut !== null) {
          l = ut;
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
                mt !== 2 && mt !== 9 || At !== t || (mt = 7), kl(t);
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
              switch (ut.tag) {
                case 26:
                  i = ut.memoizedState;
                case 5:
                case 27:
                  var f = ut;
                  if (!i || od(i)) {
                    mt = 0, Al = null;
                    var s = f.sibling;
                    if (s !== null) ut = s;
                    else {
                      var m = f.return;
                      m !== null ? (ut = m, wn(m)) : ut = null;
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
              Wc(), _t = 6;
              break t;
            default:
              throw Error(r(462));
          }
        }
        s0();
        break;
      } catch (T) {
        zo(t, T);
      }
    while (!0);
    return ae = la = null, x.H = a, x.A = u, vt = e, ut !== null ? 0 : (At = null, st = 0, dn(), _t);
  }
  function s0() {
    for (; ut !== null && !Pu(); )
      Do(ut);
  }
  function Do(t) {
    var l = lo(t.alternate, t, oe);
    t.memoizedProps = t.pendingProps, l === null ? wn(t) : ut = l;
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
          st
        );
        break;
      case 11:
        l = $r(
          e,
          l,
          l.pendingProps,
          l.type.render,
          l.ref,
          st
        );
        break;
      case 5:
        hc(l);
      default:
        ao(e, l), l = ut = Os(l, oe), l = lo(e, l, oe);
    }
    t.memoizedProps = t.pendingProps, l === null ? wn(t) : ut = l;
  }
  function Ka(t, l, e, a) {
    ae = la = null, hc(l), Ba = null, Mu = 0;
    var u = l.return;
    try {
      if (Ph(
        t,
        u,
        l,
        e,
        st
      )) {
        _t = 1, qn(
          t,
          Ol(e, t.current)
        ), ut = null;
        return;
      }
    } catch (n) {
      if (u !== null) throw ut = u, n;
      _t = 1, qn(
        t,
        Ol(e, t.current)
      ), ut = null;
      return;
    }
    l.flags & 32768 ? (ht || a === 1 ? t = !0 : Qa || (st & 536870912) !== 0 ? t = !1 : (De = t = !0, (a === 2 || a === 9 || a === 3 || a === 6) && (a = Cl.current, a !== null && a.tag === 13 && (a.flags |= 16384))), Ro(l, t)) : wn(l);
  }
  function wn(t) {
    var l = t;
    do {
      if ((l.flags & 32768) !== 0) {
        Ro(
          l,
          De
        );
        return;
      }
      t = l.return;
      var e = t0(
        l.alternate,
        l,
        oe
      );
      if (e !== null) {
        ut = e;
        return;
      }
      if (l = l.sibling, l !== null) {
        ut = l;
        return;
      }
      ut = l = t;
    } while (l !== null);
    _t === 0 && (_t = 5);
  }
  function Ro(t, l) {
    do {
      var e = l0(t.alternate, t);
      if (e !== null) {
        e.flags &= 32767, ut = e;
        return;
      }
      if (e = t.return, e !== null && (e.flags |= 32768, e.subtreeFlags = 0, e.deletions = null), !l && (t = t.sibling, t !== null)) {
        ut = t;
        return;
      }
      ut = t = e;
    } while (t !== null);
    _t = 6, ut = null;
  }
  function _o(t, l, e, a, u, n, i, f, s) {
    t.cancelPendingCommit = null;
    do
      Kn();
    while (tl !== 0);
    if ((vt & 6) !== 0) throw Error(r(327));
    if (l !== null) {
      if (l === t.current) throw Error(r(177));
      if (n = l.lanes | l.childLanes, n |= Zi, F(
        t,
        e,
        n,
        i,
        f,
        s
      ), t === At && (ut = At = null, st = 0), La = l, _e = t, Va = e, Jc = n, kc = u, jo = a, (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? (t.callbackNode = null, t.callbackPriority = 0, h0(Wl, function() {
        return Bo(), null;
      })) : (t.callbackNode = null, t.callbackPriority = 0), a = (l.flags & 13878) !== 0, (l.subtreeFlags & 13878) !== 0 || a) {
        a = x.T, x.T = null, u = H.p, H.p = 2, i = vt, vt |= 4;
        try {
          e0(t, l, e);
        } finally {
          vt = i, H.p = u, x.T = a;
        }
      }
      tl = 1, Uo(), Co(), Ho();
    }
  }
  function Uo() {
    if (tl === 1) {
      tl = 0;
      var t = _e, l = La, e = (l.flags & 13878) !== 0;
      if ((l.subtreeFlags & 13878) !== 0 || e) {
        e = x.T, x.T = null;
        var a = H.p;
        H.p = 2;
        var u = vt;
        vt |= 4;
        try {
          vo(l, t);
          var n = of, i = ps(t.containerInfo), f = n.focusedElem, s = n.selectionRange;
          if (i !== f && f && f.ownerDocument && bs(
            f.ownerDocument.documentElement,
            f
          )) {
            if (s !== null && Bi(f)) {
              var m = s.start, T = s.end;
              if (T === void 0 && (T = m), "selectionStart" in f)
                f.selectionStart = m, f.selectionEnd = Math.min(
                  T,
                  f.value.length
                );
              else {
                var z = f.ownerDocument || document, g = z && z.defaultView || window;
                if (g.getSelection) {
                  var S = g.getSelection(), W = f.textContent.length, J = Math.min(s.start, W), bt = s.end === void 0 ? J : Math.min(s.end, W);
                  !S.extend && J > bt && (i = bt, bt = J, J = i);
                  var y = Ss(
                    f,
                    J
                  ), h = Ss(
                    f,
                    bt
                  );
                  if (y && h && (S.rangeCount !== 1 || S.anchorNode !== y.node || S.anchorOffset !== y.offset || S.focusNode !== h.node || S.focusOffset !== h.offset)) {
                    var v = z.createRange();
                    v.setStart(y.node, y.offset), S.removeAllRanges(), J > bt ? (S.addRange(v), S.extend(h.node, h.offset)) : (v.setEnd(h.node, h.offset), S.addRange(v));
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
              var A = z[f];
              A.element.scrollLeft = A.left, A.element.scrollTop = A.top;
            }
          }
          ui = !!rf, of = rf = null;
        } finally {
          vt = u, H.p = a, x.T = e;
        }
      }
      t.current = l, tl = 2;
    }
  }
  function Co() {
    if (tl === 2) {
      tl = 0;
      var t = _e, l = La, e = (l.flags & 8772) !== 0;
      if ((l.subtreeFlags & 8772) !== 0 || e) {
        e = x.T, x.T = null;
        var a = H.p;
        H.p = 2;
        var u = vt;
        vt |= 4;
        try {
          ro(t, l.alternate, l);
        } finally {
          vt = u, H.p = a, x.T = e;
        }
      }
      tl = 3;
    }
  }
  function Ho() {
    if (tl === 4 || tl === 3) {
      tl = 0, Qe();
      var t = _e, l = La, e = Va, a = jo;
      (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? tl = 5 : (tl = 0, La = _e = null, qo(t, t.pendingLanes));
      var u = t.pendingLanes;
      if (u === 0 && (Re = null), ge(e), l = l.stateNode, ft && typeof ft.onCommitFiberRoot == "function")
        try {
          ft.onCommitFiberRoot(
            ol,
            l,
            void 0,
            (l.current.flags & 128) === 128
          );
        } catch {
        }
      if (a !== null) {
        l = x.T, u = H.p, H.p = 2, x.T = null;
        try {
          for (var n = t.onRecoverableError, i = 0; i < a.length; i++) {
            var f = a[i];
            n(f.value, {
              componentStack: f.stack
            });
          }
        } finally {
          x.T = l, H.p = u;
        }
      }
      (Va & 3) !== 0 && Kn(), kl(t), u = t.pendingLanes, (e & 4194090) !== 0 && (u & 42) !== 0 ? t === $c ? qu++ : (qu = 0, $c = t) : qu = 0, Bu(0);
    }
  }
  function qo(t, l) {
    (t.pooledCacheLanes &= l) === 0 && (l = t.pooledCache, l != null && (t.pooledCache = null, mu(l)));
  }
  function Kn(t) {
    return Uo(), Co(), Ho(), Bo();
  }
  function Bo() {
    if (tl !== 5) return !1;
    var t = _e, l = Jc;
    Jc = 0;
    var e = ge(Va), a = x.T, u = H.p;
    try {
      H.p = 32 > e ? 32 : e, x.T = null, e = kc, kc = null;
      var n = _e, i = Va;
      if (tl = 0, La = _e = null, Va = 0, (vt & 6) !== 0) throw Error(r(331));
      var f = vt;
      if (vt |= 4, po(n.current), go(
        n,
        n.current,
        i,
        e
      ), vt = f, Bu(0, !1), ft && typeof ft.onPostCommitFiberRoot == "function")
        try {
          ft.onPostCommitFiberRoot(ol, n);
        } catch {
        }
      return !0;
    } finally {
      H.p = u, x.T = a, qo(t, l);
    }
  }
  function Yo(t, l, e) {
    l = Ol(e, l), l = Mc(t.stateNode, l, 2), t = je(t, l, 2), t !== null && (X(t, 2), kl(t));
  }
  function jt(t, l, e) {
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
          if (typeof l.type.getDerivedStateFromError == "function" || typeof a.componentDidCatch == "function" && (Re === null || !Re.has(a))) {
            t = Ol(e, t), e = Qr(2), a = je(l, e, 2), a !== null && (Zr(
              e,
              a,
              l,
              t
            ), X(a, 2), kl(a));
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
    a !== null && a.delete(l), t.pingedLanes |= t.suspendedLanes & e, t.warmLanes &= ~e, At === t && (st & e) === e && (_t === 4 || _t === 3 && (st & 62914560) === st && 300 > bl() - Kc ? (vt & 2) === 0 && wa(t, 0) : Vc |= e, Za === st && (Za = 0)), kl(t);
  }
  function Go(t, l) {
    l === 0 && (l = me()), t = Ma(t, l), t !== null && (X(t, l), kl(t));
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
        throw Error(r(314));
    }
    a !== null && a.delete(l), Go(t, e);
  }
  function h0(t, l) {
    return Xe(t, l);
  }
  var Jn = null, Ja = null, tf = !1, kn = !1, lf = !1, fa = 0;
  function kl(t) {
    t !== Ja && t.next === null && (Ja === null ? Jn = Ja = t : Ja = Ja.next = t), kn = !0, tf || (tf = !0, v0());
  }
  function Bu(t, l) {
    if (!lf && kn) {
      lf = !0;
      do
        for (var e = !1, a = Jn; a !== null; ) {
          if (t !== 0) {
            var u = a.pendingLanes;
            if (u === 0) var n = 0;
            else {
              var i = a.suspendedLanes, f = a.pingedLanes;
              n = (1 << 31 - Pt(42 | t) + 1) - 1, n &= u & ~(i & ~f), n = n & 201326741 ? n & 201326741 | 1 : n ? n | 2 : 0;
            }
            n !== 0 && (e = !0, Lo(a, n));
          } else
            n = st, n = Le(
              a,
              a === At ? n : 0,
              a.cancelPendingCommit !== null || a.timeoutHandle !== -1
            ), (n & 3) === 0 || Ve(a, n) || (e = !0, Lo(a, n));
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
    kn = tf = !1;
    var t = 0;
    fa !== 0 && (T0() && (t = fa), fa = 0);
    for (var l = bl(), e = null, a = Jn; a !== null; ) {
      var u = a.next, n = Qo(a, l);
      n === 0 ? (a.next = null, e === null ? Jn = u : e.next = u, u === null && (Ja = e)) : (e = a, (t !== 0 || (n & 3) !== 0) && (kn = !0)), a = u;
    }
    Bu(t);
  }
  function Qo(t, l) {
    for (var e = t.suspendedLanes, a = t.pingedLanes, u = t.expirationTimes, n = t.pendingLanes & -62914561; 0 < n; ) {
      var i = 31 - Pt(n), f = 1 << i, s = u[i];
      s === -1 ? ((f & e) === 0 || (f & a) !== 0) && (u[i] = mi(f, l)) : s <= l && (t.expiredLanes |= f), n &= ~f;
    }
    if (l = At, e = st, e = Le(
      t,
      t === l ? e : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a = t.callbackNode, e === 0 || t === l && (mt === 2 || mt === 9) || t.cancelPendingCommit !== null)
      return a !== null && a !== null && Sl(a), t.callbackNode = null, t.callbackPriority = 0;
    if ((e & 3) === 0 || Ve(t, e)) {
      if (l = e & -e, l === t.callbackPriority) return l;
      switch (a !== null && Sl(a), ge(e)) {
        case 2:
        case 8:
          e = rl;
          break;
        case 32:
          e = Wl;
          break;
        case 268435456:
          e = Zl;
          break;
        default:
          e = Wl;
      }
      return a = Zo.bind(null, t), e = Xe(e, a), t.callbackPriority = l, t.callbackNode = e, l;
    }
    return a !== null && a !== null && Sl(a), t.callbackPriority = 2, t.callbackNode = null, 2;
  }
  function Zo(t, l) {
    if (tl !== 0 && tl !== 5)
      return t.callbackNode = null, t.callbackPriority = 0, null;
    var e = t.callbackNode;
    if (Kn() && t.callbackNode !== e)
      return null;
    var a = st;
    return a = Le(
      t,
      t === At ? a : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a === 0 ? null : (Eo(t, a, l), Qo(t, bl()), t.callbackNode != null && t.callbackNode === e ? Zo.bind(null, t) : null);
  }
  function Lo(t, l) {
    if (Kn()) return null;
    Eo(t, l, !0);
  }
  function v0() {
    A0(function() {
      (vt & 6) !== 0 ? Xe(
        tu,
        y0
      ) : Xo();
    });
  }
  function ef() {
    return fa === 0 && (fa = we()), fa;
  }
  function Vo(t) {
    return t == null || typeof t == "symbol" || typeof t == "boolean" ? null : typeof t == "function" ? t : un("" + t);
  }
  function wo(t, l) {
    var e = l.ownerDocument.createElement("input");
    return e.name = l.name, e.value = l.value, t.id && e.setAttribute("form", t.id), l.parentNode.insertBefore(e, l), t = new FormData(t), e.parentNode.removeChild(e), t;
  }
  function m0(t, l, e, a, u) {
    if (l === "submit" && e && e.stateNode === u) {
      var n = Vo(
        (u[al] || null).action
      ), i = a.submitter;
      i && (l = (l = i[al] || null) ? Vo(l.formAction) : i.getAttribute("formAction"), l !== null && (n = l, i = null));
      var f = new sn(
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
                if (fa !== 0) {
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
    Gl(
      g0,
      "on" + S0
    );
  }
  Gl(Ts, "onAnimationEnd"), Gl(Es, "onAnimationIteration"), Gl(As, "onAnimationStart"), Gl("dblclick", "onDoubleClick"), Gl("focusin", "onFocus"), Gl("focusout", "onBlur"), Gl(Hh, "onTransitionRun"), Gl(qh, "onTransitionStart"), Gl(Bh, "onTransitionCancel"), Gl(zs, "onTransitionEnd"), ga("onMouseEnter", ["mouseout", "mouseover"]), ga("onMouseLeave", ["mouseout", "mouseover"]), ga("onPointerEnter", ["pointerout", "pointerover"]), ga("onPointerLeave", ["pointerout", "pointerover"]), Ke(
    "onChange",
    "change click focusin focusout input keydown keyup selectionchange".split(" ")
  ), Ke(
    "onSelect",
    "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(
      " "
    )
  ), Ke("onBeforeInput", [
    "compositionend",
    "keypress",
    "textInput",
    "paste"
  ]), Ke(
    "onCompositionEnd",
    "compositionend focusout keydown keypress keyup mousedown".split(" ")
  ), Ke(
    "onCompositionStart",
    "compositionstart focusout keydown keypress keyup mousedown".split(" ")
  ), Ke(
    "onCompositionUpdate",
    "compositionupdate focusout keydown keypress keyup mousedown".split(" ")
  );
  var Yu = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(
    " "
  ), b0 = new Set(
    "beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(Yu)
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
            } catch (T) {
              Hn(T);
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
            } catch (T) {
              Hn(T);
            }
            u.currentTarget = null, n = s;
          }
      }
    }
  }
  function nt(t, l) {
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
  var $n = "_reactListening" + Math.random().toString(36).slice(2);
  function cf(t) {
    if (!t[$n]) {
      t[$n] = !0, Gf.forEach(function(e) {
        e !== "selectionchange" && (b0.has(e) || nf(e, !1, t), nf(e, !0, t));
      });
      var l = t.nodeType === 9 ? t : t.ownerDocument;
      l === null || l[$n] || (l[$n] = !0, nf("selectionchange", !1, l));
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
        u = xf;
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
      var m = n, T = zi(e), z = [];
      t: {
        var g = Ms.get(t);
        if (g !== void 0) {
          var S = sn, W = t;
          switch (t) {
            case "keypress":
              if (cn(e) === 0) break t;
            case "keydown":
            case "keyup":
              S = hh;
              break;
            case "focusin":
              W = "focus", S = _i;
              break;
            case "focusout":
              W = "blur", S = _i;
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
            case Es:
            case As:
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
              S = ph;
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
          var J = (l & 4) !== 0, bt = !J && (t === "scroll" || t === "scrollend"), y = J ? g !== null ? g + "Capture" : null : g;
          J = [];
          for (var h = m, v; h !== null; ) {
            var A = h;
            if (v = A.stateNode, A = A.tag, A !== 5 && A !== 26 && A !== 27 || v === null || y === null || (A = au(h, y), A != null && J.push(
              Gu(h, A, v)
            )), bt) break;
            h = h.return;
          }
          0 < J.length && (g = new S(
            g,
            W,
            null,
            e,
            T
          ), z.push({ event: g, listeners: J }));
        }
      }
      if ((l & 7) === 0) {
        t: {
          if (g = t === "mouseover" || t === "pointerover", S = t === "mouseout" || t === "pointerout", g && e !== Ai && (W = e.relatedTarget || e.fromElement) && (ya(W) || W[ha]))
            break t;
          if ((S || g) && (g = T.window === T ? T : (g = T.ownerDocument) ? g.defaultView || g.parentWindow : window, S ? (W = e.relatedTarget || e.toElement, S = m, W = W ? ya(W) : null, W !== null && (bt = D(W), J = W.tag, W !== bt || J !== 5 && J !== 27 && J !== 6) && (W = null)) : (S = null, W = m), S !== W)) {
            if (J = es, A = "onMouseLeave", y = "onMouseEnter", h = "mouse", (t === "pointerout" || t === "pointerover") && (J = us, A = "onPointerLeave", y = "onPointerEnter", h = "pointer"), bt = S == null ? g : eu(S), v = W == null ? g : eu(W), g = new J(
              A,
              h + "leave",
              S,
              e,
              T
            ), g.target = bt, g.relatedTarget = v, A = null, ya(T) === m && (J = new J(
              y,
              h + "enter",
              W,
              e,
              T
            ), J.target = v, J.relatedTarget = bt, A = J), bt = A, S && W)
              l: {
                for (J = S, y = W, h = 0, v = J; v; v = ka(v))
                  h++;
                for (v = 0, A = y; A; A = ka(A))
                  v++;
                for (; 0 < h - v; )
                  J = ka(J), h--;
                for (; 0 < v - h; )
                  y = ka(y), v--;
                for (; h--; ) {
                  if (J === y || y !== null && J === y.alternate)
                    break l;
                  J = ka(J), y = ka(y);
                }
                J = null;
              }
            else J = null;
            S !== null && ko(
              z,
              g,
              S,
              J,
              !1
            ), W !== null && bt !== null && ko(
              z,
              bt,
              W,
              J,
              !0
            );
          }
        }
        t: {
          if (g = m ? eu(m) : window, S = g.nodeName && g.nodeName.toLowerCase(), S === "select" || S === "input" && g.type === "file")
            var Q = ds;
          else if (rs(g))
            if (hs)
              Q = _h;
            else {
              Q = Oh;
              var et = Dh;
            }
          else
            S = g.nodeName, !S || S.toLowerCase() !== "input" || g.type !== "checkbox" && g.type !== "radio" ? m && Ei(m.elementType) && (Q = ds) : Q = Rh;
          if (Q && (Q = Q(t, m))) {
            os(
              z,
              Q,
              e,
              T
            );
            break t;
          }
          et && et(t, g, m), t === "focusout" && m && g.type === "number" && m.memoizedProps.value != null && Ti(g, "number", g.value);
        }
        switch (et = m ? eu(m) : window, t) {
          case "focusin":
            (rs(et) || et.contentEditable === "true") && (Ea = et, Yi = m, ou = null);
            break;
          case "focusout":
            ou = Yi = Ea = null;
            break;
          case "mousedown":
            Gi = !0;
            break;
          case "contextmenu":
          case "mouseup":
          case "dragend":
            Gi = !1, xs(z, e, T);
            break;
          case "selectionchange":
            if (Ch) break;
          case "keydown":
          case "keyup":
            xs(z, e, T);
        }
        var L;
        if (Ci)
          t: {
            switch (t) {
              case "compositionstart":
                var k = "onCompositionStart";
                break t;
              case "compositionend":
                k = "onCompositionEnd";
                break t;
              case "compositionupdate":
                k = "onCompositionUpdate";
                break t;
            }
            k = void 0;
          }
        else
          Ta ? fs(t, e) && (k = "onCompositionEnd") : t === "keydown" && e.keyCode === 229 && (k = "onCompositionStart");
        k && (ns && e.locale !== "ko" && (Ta || k !== "onCompositionStart" ? k === "onCompositionEnd" && Ta && (L = ts()) : (Se = T, Di = "value" in Se ? Se.value : Se.textContent, Ta = !0)), et = Wn(m, k), 0 < et.length && (k = new as(
          k,
          t,
          null,
          e,
          T
        ), z.push({ event: k, listeners: et }), L ? k.data = L : (L = ss(e), L !== null && (k.data = L)))), (L = Eh ? Ah(t, e) : zh(t, e)) && (k = Wn(m, "onBeforeInput"), 0 < k.length && (et = new as(
          "onBeforeInput",
          "beforeinput",
          null,
          e,
          T
        ), z.push({
          event: et,
          listeners: k
        }), et.data = L)), m0(
          z,
          t,
          m,
          e,
          T
        );
      }
      Ko(z, l);
    });
  }
  function Gu(t, l, e) {
    return {
      instance: t,
      listener: l,
      currentTarget: e
    };
  }
  function Wn(t, l) {
    for (var e = l + "Capture", a = []; t !== null; ) {
      var u = t, n = u.stateNode;
      if (u = u.tag, u !== 5 && u !== 26 && u !== 27 || n === null || (u = au(t, e), u != null && a.unshift(
        Gu(t, u, n)
      ), u = au(t, l), u != null && a.push(
        Gu(t, u, n)
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
      f !== 5 && f !== 26 && f !== 27 || m === null || (s = m, u ? (m = au(e, n), m != null && i.unshift(
        Gu(e, m, s)
      )) : u || (m = au(e, n), m != null && i.push(
        Gu(e, m, s)
      ))), e = e.return;
    }
    i.length !== 0 && t.push({ event: l, listeners: i });
  }
  var p0 = /\r\n?/g, x0 = /\u0000|\uFFFD/g;
  function $o(t) {
    return (typeof t == "string" ? t : "" + t).replace(p0, `
`).replace(x0, "");
  }
  function Wo(t, l) {
    return l = $o(l), $o(t) === l;
  }
  function Fn() {
  }
  function St(t, l, e, a, u, n) {
    switch (e) {
      case "children":
        typeof a == "string" ? l === "body" || l === "textarea" && a === "" || pa(t, a) : (typeof a == "number" || typeof a == "bigint") && l !== "body" && pa(t, "" + a);
        break;
      case "className":
        ln(t, "class", a);
        break;
      case "tabIndex":
        ln(t, "tabindex", a);
        break;
      case "dir":
      case "role":
      case "viewBox":
      case "width":
      case "height":
        ln(t, e, a);
        break;
      case "style":
        Ff(t, a, n);
        break;
      case "data":
        if (l !== "object") {
          ln(t, "data", a);
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
        a = un("" + a), t.setAttribute(e, a);
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
        a = un("" + a), t.setAttribute(e, a);
        break;
      case "onClick":
        a != null && (t.onclick = Fn);
        break;
      case "onScroll":
        a != null && nt("scroll", t);
        break;
      case "onScrollEnd":
        a != null && nt("scrollend", t);
        break;
      case "dangerouslySetInnerHTML":
        if (a != null) {
          if (typeof a != "object" || !("__html" in a))
            throw Error(r(61));
          if (e = a.__html, e != null) {
            if (u.children != null) throw Error(r(60));
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
        e = un("" + a), t.setAttributeNS(
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
        nt("beforetoggle", t), nt("toggle", t), tn(t, "popover", a);
        break;
      case "xlinkActuate":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:actuate",
          a
        );
        break;
      case "xlinkArcrole":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:arcrole",
          a
        );
        break;
      case "xlinkRole":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:role",
          a
        );
        break;
      case "xlinkShow":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:show",
          a
        );
        break;
      case "xlinkTitle":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:title",
          a
        );
        break;
      case "xlinkType":
        Pl(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:type",
          a
        );
        break;
      case "xmlBase":
        Pl(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:base",
          a
        );
        break;
      case "xmlLang":
        Pl(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:lang",
          a
        );
        break;
      case "xmlSpace":
        Pl(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:space",
          a
        );
        break;
      case "is":
        tn(t, "is", a);
        break;
      case "innerText":
      case "textContent":
        break;
      default:
        (!(2 < e.length) || e[0] !== "o" && e[0] !== "O" || e[1] !== "n" && e[1] !== "N") && (e = Fd.get(e) || e, tn(t, e, a));
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
            throw Error(r(61));
          if (e = a.__html, e != null) {
            if (u.children != null) throw Error(r(60));
            t.innerHTML = e;
          }
        }
        break;
      case "children":
        typeof a == "string" ? pa(t, a) : (typeof a == "number" || typeof a == "bigint") && pa(t, "" + a);
        break;
      case "onScroll":
        a != null && nt("scroll", t);
        break;
      case "onScrollEnd":
        a != null && nt("scrollend", t);
        break;
      case "onClick":
        a != null && (t.onclick = Fn);
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
            if (e[0] === "o" && e[1] === "n" && (u = e.endsWith("Capture"), l = e.slice(2, u ? e.length - 7 : void 0), n = t[al] || null, n = n != null ? n[e] : null, typeof n == "function" && t.removeEventListener(l, n, u), typeof a == "function")) {
              typeof n != "function" && n !== null && (e in t ? t[e] = null : t.hasAttribute(e) && t.removeAttribute(e)), t.addEventListener(l, a, u);
              break t;
            }
            e in t ? t[e] = a : a === !0 ? t.setAttribute(e, "") : tn(t, e, a);
          }
    }
  }
  function ll(t, l, e) {
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
        nt("error", t), nt("load", t);
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
                  throw Error(r(137, l));
                default:
                  St(t, l, n, i, e, null);
              }
          }
        u && St(t, l, "srcSet", e.srcSet, e, null), a && St(t, l, "src", e.src, e, null);
        return;
      case "input":
        nt("invalid", t);
        var f = n = i = u = null, s = null, m = null;
        for (a in e)
          if (e.hasOwnProperty(a)) {
            var T = e[a];
            if (T != null)
              switch (a) {
                case "name":
                  u = T;
                  break;
                case "type":
                  i = T;
                  break;
                case "checked":
                  s = T;
                  break;
                case "defaultChecked":
                  m = T;
                  break;
                case "value":
                  n = T;
                  break;
                case "defaultValue":
                  f = T;
                  break;
                case "children":
                case "dangerouslySetInnerHTML":
                  if (T != null)
                    throw Error(r(137, l));
                  break;
                default:
                  St(t, l, a, T, e, null);
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
        ), en(t);
        return;
      case "select":
        nt("invalid", t), a = i = n = null;
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
        nt("invalid", t), n = u = a = null;
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
                if (f != null) throw Error(r(91));
                break;
              default:
                St(t, l, i, f, e, null);
            }
        $f(t, a, u, n), en(t);
        return;
      case "option":
        for (s in e)
          e.hasOwnProperty(s) && (a = e[s], a != null) && (s === "selected" ? t.selected = a && typeof a != "function" && typeof a != "symbol" : St(t, l, s, a, e, null));
        return;
      case "dialog":
        nt("beforetoggle", t), nt("toggle", t), nt("cancel", t), nt("close", t);
        break;
      case "iframe":
      case "object":
        nt("load", t);
        break;
      case "video":
      case "audio":
        for (a = 0; a < Yu.length; a++)
          nt(Yu[a], t);
        break;
      case "image":
        nt("error", t), nt("load", t);
        break;
      case "details":
        nt("toggle", t);
        break;
      case "embed":
      case "source":
      case "link":
        nt("error", t), nt("load", t);
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
                throw Error(r(137, l));
              default:
                St(t, l, m, a, e, null);
            }
        return;
      default:
        if (Ei(l)) {
          for (T in e)
            e.hasOwnProperty(T) && (a = e[T], a !== void 0 && sf(
              t,
              l,
              T,
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
        var u = null, n = null, i = null, f = null, s = null, m = null, T = null;
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
                T = S;
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
                  throw Error(r(137, l));
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
          T,
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
                if (u != null) throw Error(r(91));
                break;
              default:
                u !== n && St(t, l, i, u, a, n);
            }
        kf(t, g, S);
        return;
      case "option":
        for (var W in e)
          g = e[W], e.hasOwnProperty(W) && g != null && !a.hasOwnProperty(W) && (W === "selected" ? t.selected = !1 : St(
            t,
            l,
            W,
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
        for (var J in e)
          g = e[J], e.hasOwnProperty(J) && g != null && !a.hasOwnProperty(J) && St(t, l, J, null, a, g);
        for (m in a)
          if (g = a[m], S = e[m], a.hasOwnProperty(m) && g !== S && (g != null || S != null))
            switch (m) {
              case "children":
              case "dangerouslySetInnerHTML":
                if (g != null)
                  throw Error(r(137, l));
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
        if (Ei(l)) {
          for (var bt in e)
            g = e[bt], e.hasOwnProperty(bt) && g !== void 0 && !a.hasOwnProperty(bt) && sf(
              t,
              l,
              bt,
              void 0,
              a,
              g
            );
          for (T in a)
            g = a[T], S = e[T], !a.hasOwnProperty(T) || g === S || g === void 0 && S === void 0 || sf(
              t,
              l,
              T,
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
  function Pn(t) {
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
  var Io = typeof setTimeout == "function" ? setTimeout : void 0, E0 = typeof clearTimeout == "function" ? clearTimeout : void 0, td = typeof Promise == "function" ? Promise : void 0, A0 = typeof queueMicrotask == "function" ? queueMicrotask : typeof td < "u" ? function(t) {
    return td.resolve(null).then(t).catch(z0);
  } : Io;
  function z0(t) {
    setTimeout(function() {
      throw t;
    });
  }
  function Ce(t) {
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
            if (e & 1 && Xu(i.documentElement), e & 2 && Xu(i.body), e & 4)
              for (e = i.head, Xu(e), i = e.firstChild; i; ) {
                var f = i.nextSibling, s = i.nodeName;
                i[lu] || s === "SCRIPT" || s === "STYLE" || s === "LINK" && i.rel.toLowerCase() === "stylesheet" || e.removeChild(i), i = f;
              }
          }
          if (u === 0) {
            t.removeChild(n), ku(l);
            return;
          }
          u--;
        } else
          e === "$" || e === "$?" || e === "$!" ? u++ : a = e.charCodeAt(0) - 48;
      else a = 0;
      e = n;
    } while (e);
    ku(l);
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
        if (!t[lu])
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
      if (t = Ql(t.nextSibling), t === null) break;
    }
    return null;
  }
  function N0(t, l, e) {
    if (l === "") return null;
    for (; t.nodeType !== 3; )
      if ((t.nodeType !== 1 || t.nodeName !== "INPUT" || t.type !== "hidden") && !e || (t = Ql(t.nextSibling), t === null)) return null;
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
  function Ql(t) {
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
    switch (l = Pn(e), t) {
      case "html":
        if (t = l.documentElement, !t) throw Error(r(452));
        return t;
      case "head":
        if (t = l.head, !t) throw Error(r(453));
        return t;
      case "body":
        if (t = l.body, !t) throw Error(r(454));
        return t;
      default:
        throw Error(r(451));
    }
  }
  function Xu(t) {
    for (var l = t.attributes; l.length; )
      t.removeAttributeNode(l[0]);
    Si(t);
  }
  var ql = /* @__PURE__ */ new Map(), ud = /* @__PURE__ */ new Set();
  function In(t) {
    return typeof t.getRootNode == "function" ? t.getRootNode() : t.nodeType === 9 ? t : t.ownerDocument;
  }
  var de = H.d;
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
    var t = de.f(), l = Vn();
    return t || l;
  }
  function R0(t) {
    var l = va(t);
    l !== null && l.tag === 5 && l.type === "form" ? Tr(l) : de.r(t);
  }
  var $a = typeof document > "u" ? null : document;
  function nd(t, l, e) {
    var a = $a;
    if (a && typeof l == "string" && l) {
      var u = Dl(l);
      u = 'link[rel="' + t + '"][href="' + u + '"]', typeof e == "string" && (u += '[crossorigin="' + e + '"]'), ud.has(u) || (ud.add(u), t = { rel: t, crossOrigin: e, href: l }, a.querySelector(u) === null && (l = a.createElement("link"), ll(l, "link", t), Kt(l), a.head.appendChild(l)));
    }
  }
  function _0(t) {
    de.D(t), nd("dns-prefetch", t, null);
  }
  function U0(t, l) {
    de.C(t, l), nd("preconnect", t, l);
  }
  function C0(t, l, e) {
    de.L(t, l, e);
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
      ), ql.set(n, t), a.querySelector(u) !== null || l === "style" && a.querySelector(Qu(n)) || l === "script" && a.querySelector(Zu(n)) || (l = a.createElement("link"), ll(l, "link", t), Kt(l), a.head.appendChild(l)));
    }
  }
  function H0(t, l) {
    de.m(t, l);
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
            if (e.querySelector(Zu(n)))
              return;
        }
        a = e.createElement("link"), ll(a, "link", t), Kt(a), e.head.appendChild(a);
      }
    }
  }
  function q0(t, l, e) {
    de.S(t, l, e);
    var a = $a;
    if (a && t) {
      var u = ma(a).hoistableStyles, n = Wa(t);
      l = l || "default";
      var i = u.get(n);
      if (!i) {
        var f = { loading: 0, preload: null };
        if (i = a.querySelector(
          Qu(n)
        ))
          f.loading = 5;
        else {
          t = C(
            { rel: "stylesheet", href: t, "data-precedence": l },
            e
          ), (e = ql.get(n)) && gf(t, e);
          var s = i = a.createElement("link");
          Kt(s), ll(s, "link", t), s._p = new Promise(function(m, T) {
            s.onload = m, s.onerror = T;
          }), s.addEventListener("load", function() {
            f.loading |= 1;
          }), s.addEventListener("error", function() {
            f.loading |= 2;
          }), f.loading |= 4, ti(i, l, a);
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
    de.X(t, l);
    var e = $a;
    if (e && t) {
      var a = ma(e).hoistableScripts, u = Fa(t), n = a.get(u);
      n || (n = e.querySelector(Zu(u)), n || (t = C({ src: t, async: !0 }, l), (l = ql.get(u)) && Sf(t, l), n = e.createElement("script"), Kt(n), ll(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function Y0(t, l) {
    de.M(t, l);
    var e = $a;
    if (e && t) {
      var a = ma(e).hoistableScripts, u = Fa(t), n = a.get(u);
      n || (n = e.querySelector(Zu(u)), n || (t = C({ src: t, async: !0, type: "module" }, l), (l = ql.get(u)) && Sf(t, l), n = e.createElement("script"), Kt(n), ll(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function id(t, l, e, a) {
    var u = (u = $.current) ? In(u) : null;
    if (!u) throw Error(r(446));
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
            Qu(t)
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
            throw Error(r(528, ""));
          return i;
        }
        if (l && a !== null)
          throw Error(r(529, ""));
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
        throw Error(r(444, t));
    }
  }
  function Wa(t) {
    return 'href="' + Dl(t) + '"';
  }
  function Qu(t) {
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
    }), ll(l, "link", e), Kt(l), t.head.appendChild(l));
  }
  function Fa(t) {
    return '[src="' + Dl(t) + '"]';
  }
  function Zu(t) {
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
            return l.instance = a, Kt(a), a;
          var u = C({}, e, {
            "data-href": e.href,
            "data-precedence": e.precedence,
            href: null,
            precedence: null
          });
          return a = (t.ownerDocument || t).createElement(
            "style"
          ), Kt(a), ll(a, "style", u), ti(a, e.precedence, t), l.instance = a;
        case "stylesheet":
          u = Wa(e.href);
          var n = t.querySelector(
            Qu(u)
          );
          if (n)
            return l.state.loading |= 4, l.instance = n, Kt(n), n;
          a = cd(e), (u = ql.get(u)) && gf(a, u), n = (t.ownerDocument || t).createElement("link"), Kt(n);
          var i = n;
          return i._p = new Promise(function(f, s) {
            i.onload = f, i.onerror = s;
          }), ll(n, "link", a), l.state.loading |= 4, ti(n, e.precedence, t), l.instance = n;
        case "script":
          return n = Fa(e.src), (u = t.querySelector(
            Zu(n)
          )) ? (l.instance = u, Kt(u), u) : (a = e, (u = ql.get(n)) && (a = C({}, e), Sf(a, u)), t = t.ownerDocument || t, u = t.createElement("script"), Kt(u), ll(u, "link", a), t.head.appendChild(u), l.instance = u);
        case "void":
          return null;
        default:
          throw Error(r(443, l.type));
      }
    else
      l.type === "stylesheet" && (l.state.loading & 4) === 0 && (a = l.instance, l.state.loading |= 4, ti(a, e.precedence, t));
    return l.instance;
  }
  function ti(t, l, e) {
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
  var li = null;
  function sd(t, l, e) {
    if (li === null) {
      var a = /* @__PURE__ */ new Map(), u = li = /* @__PURE__ */ new Map();
      u.set(e, a);
    } else
      u = li, a = u.get(e), a || (a = /* @__PURE__ */ new Map(), u.set(e, a));
    if (a.has(t)) return a;
    for (a.set(t, null), e = e.getElementsByTagName(t), u = 0; u < e.length; u++) {
      var n = e[u];
      if (!(n[lu] || n[Yt] || t === "link" && n.getAttribute("rel") === "stylesheet") && n.namespaceURI !== "http://www.w3.org/2000/svg") {
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
  var Lu = null;
  function Q0() {
  }
  function Z0(t, l, e) {
    if (Lu === null) throw Error(r(475));
    var a = Lu;
    if (l.type === "stylesheet" && (typeof e.media != "string" || matchMedia(e.media).matches !== !1) && (l.state.loading & 4) === 0) {
      if (l.instance === null) {
        var u = Wa(e.href), n = t.querySelector(
          Qu(u)
        );
        if (n) {
          t = n._p, t !== null && typeof t == "object" && typeof t.then == "function" && (a.count++, a = ei.bind(a), t.then(a, a)), l.state.loading |= 4, l.instance = n, Kt(n);
          return;
        }
        n = t.ownerDocument || t, e = cd(e), (u = ql.get(u)) && gf(e, u), n = n.createElement("link"), Kt(n);
        var i = n;
        i._p = new Promise(function(f, s) {
          i.onload = f, i.onerror = s;
        }), ll(n, "link", e), l.instance = n;
      }
      a.stylesheets === null && (a.stylesheets = /* @__PURE__ */ new Map()), a.stylesheets.set(l, t), (t = l.state.preload) && (l.state.loading & 3) === 0 && (a.count++, l = ei.bind(a), t.addEventListener("load", l), t.addEventListener("error", l));
    }
  }
  function L0() {
    if (Lu === null) throw Error(r(475));
    var t = Lu;
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
  function ei() {
    if (this.count--, this.count === 0) {
      if (this.stylesheets) bf(this, this.stylesheets);
      else if (this.unsuspend) {
        var t = this.unsuspend;
        this.unsuspend = null, t();
      }
    }
  }
  var ai = null;
  function bf(t, l) {
    t.stylesheets = null, t.unsuspend !== null && (t.count++, ai = /* @__PURE__ */ new Map(), l.forEach(V0, t), ai = null, ei.call(t));
  }
  function V0(t, l) {
    if (!(l.state.loading & 4)) {
      var e = ai.get(t);
      if (e) var a = e.get(null);
      else {
        e = /* @__PURE__ */ new Map(), ai.set(t, e);
        for (var u = t.querySelectorAll(
          "link[data-precedence],style[data-precedence]"
        ), n = 0; n < u.length; n++) {
          var i = u[n];
          (i.nodeName === "LINK" || i.getAttribute("media") !== "not all") && (e.set(i.dataset.precedence, i), a = i);
        }
        a && e.set(null, a);
      }
      u = l.instance, i = u.getAttribute("data-precedence"), n = e.get(i) || a, n === a && e.set(null, u), e.set(i, u), this.count++, a = ei.bind(this), u.addEventListener("load", a), u.addEventListener("error", a), n ? n.parentNode.insertBefore(u, n.nextSibling) : (t = t.nodeType === 9 ? t.head : t, t.insertBefore(u, t.firstChild)), l.state.loading |= 4;
    }
  }
  var Vu = {
    $$typeof: Et,
    Provider: null,
    Consumer: null,
    _currentValue: K,
    _currentValue2: K,
    _threadCount: 0
  };
  function w0(t, l, e, a, u, n, i, f) {
    this.tag = 1, this.containerInfo = t, this.pingCache = this.current = this.pendingChildren = null, this.timeoutHandle = -1, this.callbackNode = this.next = this.pendingContext = this.context = this.cancelPendingCommit = null, this.callbackPriority = 0, this.expirationTimes = E(-1), this.entangledLanes = this.shellSuspendCounter = this.errorRecoveryDisabledLanes = this.expiredLanes = this.warmLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0, this.entanglements = E(0), this.hiddenUpdates = E(null), this.identifierPrefix = a, this.onUncaughtError = u, this.onCaughtError = n, this.onRecoverableError = i, this.pooledCache = null, this.pooledCacheLanes = 0, this.formState = f, this.incompleteTransitions = /* @__PURE__ */ new Map();
  }
  function dd(t, l, e, a, u, n, i, f, s, m, T, z) {
    return t = new w0(
      t,
      l,
      e,
      i,
      f,
      s,
      m,
      z
    ), l = 1, n === !0 && (l |= 24), n = jl(3, null, null, l), t.current = n, n.stateNode = t, l = Ii(), l.refCount++, t.pooledCache = l, l.refCount++, n.memoizedState = {
      element: a,
      isDehydrated: e,
      cache: l
    }, ac(n), t;
  }
  function hd(t) {
    return t ? (t = Na, t) : Na;
  }
  function yd(t, l, e, a, u, n) {
    u = hd(u), a.context === null ? a.context = u : a.pendingContext = u, a = xe(l), a.payload = { element: e }, n = n === void 0 ? null : n, n !== null && (a.callback = n), e = je(t, a, l), e !== null && (Ml(e, t, l), pu(e, t, l));
  }
  function vd(t, l) {
    if (t = t.memoizedState, t !== null && t.dehydrated !== null) {
      var e = t.retryLane;
      t.retryLane = e !== 0 && e < l ? e : l;
    }
  }
  function pf(t, l) {
    vd(t, l), (t = t.alternate) && vd(t, l);
  }
  function md(t) {
    if (t.tag === 13) {
      var l = Ma(t, 67108864);
      l !== null && Ml(l, t, 67108864), pf(t, 67108864);
    }
  }
  var ui = !0;
  function K0(t, l, e, a) {
    var u = x.T;
    x.T = null;
    var n = H.p;
    try {
      H.p = 2, xf(t, l, e, a);
    } finally {
      H.p = n, x.T = u;
    }
  }
  function J0(t, l, e, a) {
    var u = x.T;
    x.T = null;
    var n = H.p;
    try {
      H.p = 8, xf(t, l, e, a);
    } finally {
      H.p = n, x.T = u;
    }
  }
  function xf(t, l, e, a) {
    if (ui) {
      var u = jf(a);
      if (u === null)
        ff(
          t,
          l,
          a,
          ni,
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
                  var i = Ll(n.pendingLanes);
                  if (i !== 0) {
                    var f = n;
                    for (f.pendingLanes |= 2, f.entangledLanes |= 2; i; ) {
                      var s = 1 << 31 - Pt(i);
                      f.entanglements[1] |= s, i &= ~s;
                    }
                    kl(n), (vt & 6) === 0 && (Zn = bl() + 500, Bu(0));
                  }
                }
                break;
              case 13:
                f = Ma(n, 2), f !== null && Ml(f, n, 2), Vn(), pf(n, 2);
            }
          if (n = jf(a), n === null && ff(
            t,
            l,
            a,
            ni,
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
  var ni = null;
  function Tf(t) {
    if (ni = null, t = ya(t), t !== null) {
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
    return ni = t, null;
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
        switch (vi()) {
          case tu:
            return 2;
          case rl:
            return 8;
          case Wl:
          case ye:
            return 32;
          case Zl:
            return 268435456;
          default:
            return 32;
        }
      default:
        return 32;
    }
  }
  var Ef = !1, He = null, qe = null, Be = null, wu = /* @__PURE__ */ new Map(), Ku = /* @__PURE__ */ new Map(), Ye = [], k0 = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(
    " "
  );
  function Sd(t, l) {
    switch (t) {
      case "focusin":
      case "focusout":
        He = null;
        break;
      case "dragenter":
      case "dragleave":
        qe = null;
        break;
      case "mouseover":
      case "mouseout":
        Be = null;
        break;
      case "pointerover":
      case "pointerout":
        wu.delete(l.pointerId);
        break;
      case "gotpointercapture":
      case "lostpointercapture":
        Ku.delete(l.pointerId);
    }
  }
  function Ju(t, l, e, a, u, n) {
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
        return He = Ju(
          He,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "dragenter":
        return qe = Ju(
          qe,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "mouseover":
        return Be = Ju(
          Be,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "pointerover":
        var n = u.pointerId;
        return wu.set(
          n,
          Ju(
            wu.get(n) || null,
            t,
            l,
            e,
            a,
            u
          )
        ), !0;
      case "gotpointercapture":
        return n = u.pointerId, Ku.set(
          n,
          Ju(
            Ku.get(n) || null,
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
            t.blockedOn = l, Fl(t.priority, function() {
              if (e.tag === 13) {
                var a = zl();
                a = Lt(a);
                var u = Ma(e, a);
                u !== null && Ml(u, e, a), pf(e, a);
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
  function ii(t) {
    if (t.blockedOn !== null) return !1;
    for (var l = t.targetContainers; 0 < l.length; ) {
      var e = jf(t.nativeEvent);
      if (e === null) {
        e = t.nativeEvent;
        var a = new e.constructor(
          e.type,
          e
        );
        Ai = a, e.target.dispatchEvent(a), Ai = null;
      } else
        return l = va(e), l !== null && md(l), t.blockedOn = e, !1;
      l.shift();
    }
    return !0;
  }
  function pd(t, l, e) {
    ii(t) && e.delete(l);
  }
  function W0() {
    Ef = !1, He !== null && ii(He) && (He = null), qe !== null && ii(qe) && (qe = null), Be !== null && ii(Be) && (Be = null), wu.forEach(pd), Ku.forEach(pd);
  }
  function ci(t, l) {
    t.blockedOn === l && (t.blockedOn = null, Ef || (Ef = !0, o.unstable_scheduleCallback(
      o.unstable_NormalPriority,
      W0
    )));
  }
  var fi = null;
  function xd(t) {
    fi !== t && (fi = t, o.unstable_scheduleCallback(
      o.unstable_NormalPriority,
      function() {
        fi === t && (fi = null);
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
  function ku(t) {
    function l(s) {
      return ci(s, t);
    }
    He !== null && ci(He, t), qe !== null && ci(qe, t), Be !== null && ci(Be, t), wu.forEach(l), Ku.forEach(l);
    for (var e = 0; e < Ye.length; e++) {
      var a = Ye[e];
      a.blockedOn === t && (a.blockedOn = null);
    }
    for (; 0 < Ye.length && (e = Ye[0], e.blockedOn === null); )
      bd(e), e.blockedOn === null && Ye.shift();
    if (e = (t.ownerDocument || t).$$reactFormReplay, e != null)
      for (a = 0; a < e.length; a += 3) {
        var u = e[a], n = e[a + 1], i = u[al] || null;
        if (typeof n == "function")
          i || xd(e);
        else if (i) {
          var f = null;
          if (n && n.hasAttribute("formAction")) {
            if (u = n, i = n[al] || null)
              f = i.formAction;
            else if (Tf(u) !== null) continue;
          } else f = i.action;
          typeof f == "function" ? e[a + 1] = f : (e.splice(a, 3), a -= 3), xd(e);
        }
      }
  }
  function Af(t) {
    this._internalRoot = t;
  }
  si.prototype.render = Af.prototype.render = function(t) {
    var l = this._internalRoot;
    if (l === null) throw Error(r(409));
    var e = l.current, a = zl();
    yd(e, a, t, l, null, null);
  }, si.prototype.unmount = Af.prototype.unmount = function() {
    var t = this._internalRoot;
    if (t !== null) {
      this._internalRoot = null;
      var l = t.containerInfo;
      yd(t.current, 2, null, t, null, null), Vn(), l[ha] = null;
    }
  };
  function si(t) {
    this._internalRoot = t;
  }
  si.prototype.unstable_scheduleHydration = function(t) {
    if (t) {
      var l = dl();
      t = { blockedOn: null, target: t, priority: l };
      for (var e = 0; e < Ye.length && l !== 0 && l < Ye[e].priority; e++) ;
      Ye.splice(e, 0, t), e === 0 && bd(t);
    }
  };
  var jd = b.version;
  if (jd !== "19.1.1")
    throw Error(
      r(
        527,
        jd,
        "19.1.1"
      )
    );
  H.findDOMNode = function(t) {
    var l = t._reactInternals;
    if (l === void 0)
      throw typeof t.render == "function" ? Error(r(188)) : (t = Object.keys(t).join(","), Error(r(268, t)));
    return t = M(l), t = t !== null ? p(t) : null, t = t === null ? null : t.stateNode, t;
  };
  var F0 = {
    bundleType: 0,
    version: "19.1.1",
    rendererPackageName: "react-dom",
    currentDispatcherRef: x,
    reconcilerVersion: "19.1.1"
  };
  if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u") {
    var ri = __REACT_DEVTOOLS_GLOBAL_HOOK__;
    if (!ri.isDisabled && ri.supportsFiber)
      try {
        ol = ri.inject(
          F0
        ), ft = ri;
      } catch {
      }
  }
  return Wu.createRoot = function(t, l) {
    if (!_(t)) throw Error(r(299));
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
    ), t[ha] = l.current, cf(t), new Af(l);
  }, Wu.hydrateRoot = function(t, l, e) {
    if (!_(t)) throw Error(r(299));
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
    ), l.context = hd(null), e = l.current, a = zl(), a = Lt(a), u = xe(a), u.callback = null, je(e, u, a), e = a, l.current.lanes = e, X(l, e), kl(l), t[ha] = l.current, cf(t), new si(l);
  }, Wu.version = "19.1.1", Wu;
}
var _d;
function c1() {
  if (_d) return Mf.exports;
  _d = 1;
  function o() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o);
      } catch (b) {
        console.error(b);
      }
  }
  return o(), Mf.exports = i1(), Mf.exports;
}
var f1 = c1(), R = Uf();
const s1 = (o) => o.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase(), Bd = (...o) => o.filter((b, j, r) => !!b && b.trim() !== "" && r.indexOf(b) === j).join(" ").trim();
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
    color: o = "currentColor",
    size: b = 24,
    strokeWidth: j = 2,
    absoluteStrokeWidth: r,
    className: _ = "",
    children: D,
    iconNode: U,
    ...O
  }, M) => R.createElement(
    "svg",
    {
      ref: M,
      ...r1,
      width: b,
      height: b,
      stroke: o,
      strokeWidth: r ? Number(j) * 24 / Number(b) : j,
      className: Bd("lucide", _),
      ...O
    },
    [
      ...U.map(([p, C]) => R.createElement(p, C)),
      ...Array.isArray(D) ? D : [D]
    ]
  )
);
const Tt = (o, b) => {
  const j = R.forwardRef(
    ({ className: r, ..._ }, D) => R.createElement(o1, {
      ref: D,
      iconNode: b,
      className: Bd(`lucide-${s1(o)}`, r),
      ..._
    })
  );
  return j.displayName = `${o}`, j;
};
const _f = Tt("AudioLines", [
  ["path", { d: "M2 10v3", key: "1fnikh" }],
  ["path", { d: "M6 6v11", key: "11sgs0" }],
  ["path", { d: "M10 3v18", key: "yhl04a" }],
  ["path", { d: "M14 8v7", key: "3a1oy3" }],
  ["path", { d: "M18 5v13", key: "123xd1" }],
  ["path", { d: "M22 10v3", key: "154ddg" }]
]);
const Fu = Tt("Check", [["path", { d: "M20 6 9 17l-5-5", key: "1gmf2c" }]]);
const hi = Tt("ChevronDown", [
  ["path", { d: "m6 9 6 6 6-6", key: "qrunsl" }]
]);
const Cf = Tt("ChevronRight", [
  ["path", { d: "m9 18 6-6-6-6", key: "mthhwq" }]
]);
const d1 = Tt("ChevronUp", [["path", { d: "m18 15-6-6-6 6", key: "153udz" }]]);
const Yd = Tt("CircleCheck", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["path", { d: "m9 12 2 2 4-4", key: "dzmm74" }]
]);
const h1 = Tt("Circle", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }]
]);
const y1 = Tt("Clock3", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["polyline", { points: "12 6 12 12 16.5 12", key: "1aq6pp" }]
]);
const v1 = Tt("CodeXml", [
  ["path", { d: "m18 16 4-4-4-4", key: "1inbqp" }],
  ["path", { d: "m6 8-4 4 4 4", key: "15zrgr" }],
  ["path", { d: "m14.5 4-5 16", key: "e7oirm" }]
]);
const m1 = Tt("Ellipsis", [
  ["circle", { cx: "12", cy: "12", r: "1", key: "41hilf" }],
  ["circle", { cx: "19", cy: "12", r: "1", key: "1wjl8i" }],
  ["circle", { cx: "5", cy: "12", r: "1", key: "1pcz8c" }]
]);
const Hf = Tt("FileAudio", [
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
const qf = Tt("FolderGit2", [
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
const he = Tt("LoaderCircle", [
  ["path", { d: "M21 12a9 9 0 1 1-6.219-8.56", key: "13zald" }]
]);
const g1 = Tt("Mic", [
  ["path", { d: "M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z", key: "131961" }],
  ["path", { d: "M19 10v2a7 7 0 0 1-14 0v-2", key: "1vc78b" }],
  ["line", { x1: "12", x2: "12", y1: "19", y2: "22", key: "x3vr5v" }]
]);
const S1 = Tt("Pause", [
  ["rect", { x: "14", y: "4", width: "4", height: "16", rx: "1", key: "zuxfzm" }],
  ["rect", { x: "6", y: "4", width: "4", height: "16", rx: "1", key: "1okwgv" }]
]);
const oi = Tt("PencilLine", [
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
const b1 = Tt("Play", [
  ["polygon", { points: "6 3 20 12 6 21 6 3", key: "1oa8hb" }]
]);
const di = Tt("Plus", [
  ["path", { d: "M5 12h14", key: "1ays0h" }],
  ["path", { d: "M12 5v14", key: "s699le" }]
]);
const Gd = Tt("Radio", [
  ["path", { d: "M4.9 19.1C1 15.2 1 8.8 4.9 4.9", key: "1vaf9d" }],
  ["path", { d: "M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5", key: "u1ii0m" }],
  ["circle", { cx: "12", cy: "12", r: "2", key: "1c9p78" }],
  ["path", { d: "M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5", key: "1j5fej" }],
  ["path", { d: "M19.1 4.9C23 8.8 23 15.1 19.1 19", key: "10b0cb" }]
]);
const p1 = Tt("RefreshCw", [
  ["path", { d: "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8", key: "v9h5vc" }],
  ["path", { d: "M21 3v5h-5", key: "1q7to0" }],
  ["path", { d: "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16", key: "3uifl3" }],
  ["path", { d: "M8 16H3v5", key: "1cv678" }]
]);
const x1 = Tt("Search", [
  ["circle", { cx: "11", cy: "11", r: "8", key: "4ej97u" }],
  ["path", { d: "m21 21-4.3-4.3", key: "1qie3q" }]
]);
const j1 = Tt("Send", [
  [
    "path",
    {
      d: "M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z",
      key: "1ffxy3"
    }
  ],
  ["path", { d: "m21.854 2.147-10.94 10.939", key: "12cjpa" }]
]);
const T1 = Tt("Settings2", [
  ["path", { d: "M20 7h-9", key: "3s1dr2" }],
  ["path", { d: "M14 17H5", key: "gfn3mx" }],
  ["circle", { cx: "17", cy: "17", r: "3", key: "18b49y" }],
  ["circle", { cx: "7", cy: "7", r: "3", key: "dfmy0x" }]
]);
const Ud = Tt("Sparkles", [
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
const E1 = Tt("Trash2", [
  ["path", { d: "M3 6h18", key: "d0wm0j" }],
  ["path", { d: "M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6", key: "4alrt4" }],
  ["path", { d: "M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2", key: "v07s0e" }],
  ["line", { x1: "10", x2: "10", y1: "11", y2: "17", key: "1uufr5" }],
  ["line", { x1: "14", x2: "14", y1: "11", y2: "17", key: "xtxkd" }]
]);
const A1 = Tt("X", [
  ["path", { d: "M18 6 6 18", key: "1bl5f8" }],
  ["path", { d: "m6 6 12 12", key: "d8bk6v" }]
]);
const z1 = Tt("Zap", [
  [
    "path",
    {
      d: "M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z",
      key: "1xq2db"
    }
  ]
]);
let yi = async () => {
  throw Error("请从已连接的设备打开 SaySo");
};
function M1(o) {
  yi = o;
}
function N1(o, b, j) {
  return yi(o, b, j);
}
function D1(o) {
  return Uint8Array.from(atob(o), (b) => b.charCodeAt(0));
}
let Cd = Promise.resolve();
function O1(o, b, j) {
  const r = yi, _ = Cd.catch(() => {
  }).then(async () => {
    const { id: D } = await r("POST", "/api/audio/begin", {
      ...j,
      sessionId: o,
      sizeBytes: b.size,
      mimeType: b.type || "audio/webm"
    });
    try {
      for (let U = 0; U < b.size; U += 16e3) {
        const O = new Uint8Array(
          await b.slice(U, U + 16e3).arrayBuffer()
        );
        await r("POST", `/api/audio/${D}/chunk`, {
          offset: U,
          data: btoa(String.fromCharCode(...O))
        });
      }
      return await r("POST", `/api/audio/${D}/finish`);
    } catch (U) {
      throw await r("POST", `/api/audio/${D}/abort`).catch(
        () => {
        }
      ), U;
    }
  });
  return Cd = _, _;
}
async function R1(o, b) {
  const j = yi, r = [];
  let _ = 0, D = 0, U = "";
  do {
    if (b.aborted) throw Error("已取消");
    const O = await j("POST", `/api/audio/${o}/read`, {
      offset: _
    });
    if (!O.size || O.size > 64 * 1024 * 1024)
      throw Error("录音大小不合法");
    const M = D1(O.data);
    if (!M.length) throw Error("录音数据不完整");
    r.push(M), _ += M.length, D = O.size, U = O.mimeType;
  } while (_ < D);
  if (b.aborted) throw Error("已取消");
  return URL.createObjectURL(new Blob(r, { type: U }));
}
async function fl(o, b) {
  return N1(
    b?.method || "GET",
    o,
    b?.body ? JSON.parse(String(b.body)) : void 0
  );
}
const qt = {
  assistants: () => fl("/api/assistants"),
  cancelRun: (o) => fl(`/api/runs/${o}/cancel`, { method: "POST" }),
  state: () => fl("/api/state"),
  updateSettings: (o) => fl("/api/settings", {
    method: "PATCH",
    body: JSON.stringify(o)
  }),
  createProject: (o) => fl("/api/projects", {
    method: "POST",
    body: JSON.stringify(o)
  }),
  updateProject: (o, b) => fl(`/api/projects/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  createSession: (o, b) => fl("/api/sessions", {
    method: "POST",
    body: JSON.stringify({ projectId: o, title: b })
  }),
  updateSession: (o, b) => fl(`/api/sessions/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  deleteSession: (o) => fl(`/api/sessions/${o}`, { method: "DELETE" }),
  addUtterance: (o, b) => fl(`/api/sessions/${o}/utterances`, {
    method: "POST",
    body: JSON.stringify(b)
  }),
  extractActions: (o) => fl(`/api/sessions/${o}/extract-actions`, { method: "POST" }),
  createAction: (o, b) => fl(`/api/sessions/${o}/actions`, {
    method: "POST",
    body: JSON.stringify(b)
  }),
  updateAction: (o, b) => fl(`/api/actions/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  deleteAction: (o) => fl(`/api/actions/${o}`, { method: "DELETE" }),
  dispatchAction: (o, b) => fl(`/api/actions/${o}/dispatch`, { method: "POST", body: JSON.stringify(b ? { target: b } : {}) }),
  uploadAudio: O1
};
function _1({ action: o, onClose: b, onSend: j }) {
  const [r, _] = R.useState([]), [D, U] = R.useState(""), [O, M] = R.useState(!1), [p, C] = R.useState("");
  return R.useEffect(() => {
    let G = !0;
    return qt.assistants().then((Z) => {
      G && _(Z.assistants.filter((tt) => tt.scope !== "account"));
    }).catch((Z) => {
      G && C(Z.message);
    }), () => {
      G = !1;
    };
  }, []), /* @__PURE__ */ c.jsx("div", { className: "modal-backdrop", children: /* @__PURE__ */ c.jsxs("section", { className: "modal", role: "dialog", "aria-modal": "true", "aria-label": "选择执行管家", children: [
    /* @__PURE__ */ c.jsxs("header", { className: "modal-header", children: [
      /* @__PURE__ */ c.jsx("h2", { children: "选择执行管家" }),
      /* @__PURE__ */ c.jsx("button", { onClick: b, disabled: O, "aria-label": "关闭", children: "×" })
    ] }),
    /* @__PURE__ */ c.jsx("p", { children: o.title }),
    /* @__PURE__ */ c.jsx("p", { children: "Action 将交给选定管家，完成后回报 CEO；讨论和录音保存在 CEO 主机。" }),
    /* @__PURE__ */ c.jsxs("label", { children: [
      "执行目标",
      /* @__PURE__ */ c.jsxs("select", { "aria-label": "执行目标", value: D, onChange: (G) => U(G.target.value), disabled: O, children: [
        /* @__PURE__ */ c.jsx("option", { value: "", children: "请选择客户端管家或项目管家" }),
        r.map((G, Z) => /* @__PURE__ */ c.jsxs("option", { value: String(Z), children: [
          G.name || G.projectId || G.clientId,
          G.online === !1 ? "（离线，等待重连）" : ""
        ] }, Z))
      ] })
    ] }),
    p && /* @__PURE__ */ c.jsx("p", { role: "alert", children: p }),
    /* @__PURE__ */ c.jsxs("footer", { className: "modal-footer", children: [
      /* @__PURE__ */ c.jsx("button", { onClick: b, disabled: O, children: "取消" }),
      /* @__PURE__ */ c.jsx("button", { className: "primary-button", disabled: O || D === "", onClick: async () => {
        M(!0), C("");
        try {
          await j(r[Number(D)]);
        } catch (G) {
          C(G instanceof Error ? G.message : "派工失败");
        } finally {
          M(!1);
        }
      }, children: O ? "正在加入派工队列…" : "确认派工" })
    ] })
  ] }) });
}
function Bf(o) {
  return o.trim().toLocaleLowerCase();
}
function U1(o, b) {
  const j = Bf(b);
  return j.length > 0 && o.toLocaleLowerCase().includes(j);
}
function C1(o, b) {
  const j = Bf(b);
  if (!j) return [{ text: o, match: !1 }];
  const r = o.toLocaleLowerCase(), _ = [];
  let D = 0;
  for (; D < o.length; ) {
    const U = r.indexOf(j, D);
    if (U === -1) {
      _.push({ text: o.slice(D), match: !1 });
      break;
    }
    U > D && _.push({ text: o.slice(D, U), match: !1 }), _.push({
      text: o.slice(U, U + j.length),
      match: !0
    }), D = U + j.length;
  }
  return _.length ? _ : [{ text: o, match: !1 }];
}
const H1 = 600 * 1e3, q1 = {
  analysisProvider: "codex",
  analysisModel: "default",
  analysisIntervalMinutes: 2,
  executor: "codex",
  executionModel: "default"
}, B1 = {
  projects: [],
  sessions: [],
  actions: [],
  runs: [],
  settings: q1
}, Y1 = [
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini（需账号支持）" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "default", label: "Codex 默认模型（推荐）" }
], G1 = [
  { value: "default", label: "Codex 默认模型（推荐）" },
  { value: "gpt-5.6-sol", label: "GPT-5.6 Sol" },
  { value: "gpt-5.6-terra", label: "GPT-5.6 Terra" },
  { value: "gpt-5.6-luna", label: "GPT-5.6 Luna" },
  { value: "gpt-5.4-mini", label: "GPT-5.4 mini" }
];
function Xd(o) {
  const b = Math.max(0, Math.floor(o / 1e3));
  return `${String(Math.floor(b / 60)).padStart(2, "0")}:${String(b % 60).padStart(2, "0")}`;
}
function Hd(o) {
  const b = new Date(o);
  return Number.isNaN(b.getTime()) ? o : `${b.getFullYear()}-${b.getMonth()}-${b.getDate()}-${b.getHours()}-${b.getMinutes()}`;
}
function X1(o) {
  const b = new Date(o);
  return Number.isNaN(b.getTime()) ? "--:--" : `${String(b.getHours()).padStart(2, "0")}:${String(b.getMinutes()).padStart(2, "0")}`;
}
function Q1(o) {
  return o ? o < 1024 * 1024 ? `${Math.max(1, Math.round(o / 1024))} KB` : `${(o / 1024 / 1024).toFixed(1)} MB` : "大小未知";
}
function qd(o) {
  const b = new Date(o), j = /* @__PURE__ */ new Date();
  if (b.toDateString() === j.toDateString()) return "今天";
  const r = new Date(j);
  return r.setDate(j.getDate() - 1), b.toDateString() === r.toDateString() ? "昨天" : `${b.getMonth() + 1}月${b.getDate()}日`;
}
function sa({
  title: o,
  description: b,
  children: j,
  onClose: r,
  wide: _ = !1
}) {
  return /* @__PURE__ */ c.jsx(
    "div",
    {
      className: "modal-backdrop",
      onMouseDown: (D) => D.target === D.currentTarget && r(),
      children: /* @__PURE__ */ c.jsxs(
        "section",
        {
          className: `modal ${_ ? "modal-wide" : ""}`,
          role: "dialog",
          "aria-modal": "true",
          children: [
            /* @__PURE__ */ c.jsxs("div", { className: "modal-heading", children: [
              /* @__PURE__ */ c.jsxs("div", { children: [
                /* @__PURE__ */ c.jsx("h2", { children: o }),
                b && /* @__PURE__ */ c.jsx("p", { children: b })
              ] }),
              /* @__PURE__ */ c.jsx("button", { className: "icon-button", onClick: r, "aria-label": "关闭", children: /* @__PURE__ */ c.jsx(A1, { size: 18 }) })
            ] }),
            j
          ]
        }
      )
    }
  );
}
function Qd({ status: o }) {
  const b = {
    draft: "待完善",
    ready: "Ready",
    running: "Web 会话工作中",
    done: "已完成",
    failed: "执行失败"
  }[o];
  return /* @__PURE__ */ c.jsxs("span", { className: `status-pill status-${o}`, children: [
    o === "running" && /* @__PURE__ */ c.jsx(he, { size: 12 }),
    " ",
    b
  ] });
}
function Zd(o, b) {
  const [j, r] = R.useState([]), _ = o.join("\0");
  return R.useEffect(() => {
    if (!b) {
      r([]);
      return;
    }
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      r(o);
      return;
    }
    let D = !1, U = 0, O = 0, M;
    r([]);
    const p = () => {
      if (D) return;
      const C = Array.from(o[U]);
      if (O < C.length) {
        O += 1;
        const G = C.slice(0, O).join("");
        r((Z) => {
          const tt = Z.slice();
          return tt[U] = G, tt;
        }), M = window.setTimeout(p, 34);
        return;
      }
      if (U < o.length - 1) {
        U += 1, O = 0, M = window.setTimeout(p, 360);
        return;
      }
      M = window.setTimeout(() => {
        U = 0, O = 0, r([]), M = window.setTimeout(p, 240);
      }, 1600);
    };
    return M = window.setTimeout(p, 160), () => {
      D = !0, M !== void 0 && window.clearTimeout(M);
    };
  }, [b, _]), j;
}
function Z1({
  onClose: o,
  deviceLabel: b,
  scope: j,
  registerClose: r
}) {
  const [_, D] = R.useState(!1), [U, O] = R.useState(null), [M, p] = R.useState([]), C = R.useRef([]), G = R.useRef(/* @__PURE__ */ new Set()), Z = R.useRef(async () => {
  }), [tt, B] = R.useState(""), [yt, Nt] = R.useState(!1), [Wt, pt] = R.useState(
    "discussion"
  ), [Et, Bt] = R.useState(!1);
  R.useEffect(() => {
    const E = (F) => {
      (rl.current || G.current.size) && (F.preventDefault(), F.returnValue = "");
    }, X = () => {
      Z.current().catch(() => {
      }).finally(() => ft("设备连接已断开，未上传的录音可下载保存。"));
    };
    return window.addEventListener("beforeunload", E), window.addEventListener("remote-agent-disconnected", X), () => {
      ye.current = !0;
      try {
        Wl.current?.stop();
      } catch {
      }
      Zl.current?.getTracks().forEach((F) => F.stop()), pl.current && window.clearTimeout(pl.current), C.current.forEach((F) => URL.revokeObjectURL(F.url)), window.removeEventListener("beforeunload", E), window.removeEventListener("remote-agent-disconnected", X);
    };
  }, []);
  async function P() {
    D(!0);
    try {
      return await Z.current(), await Promise.allSettled([...G.current]), C.current.length && !confirm("仍有未上传录音，请先下载保存。确定离开吗？") ? !1 : (o(), !0);
    } catch (E) {
      return ft(
        E instanceof Error ? E.message : "正在保存，请稍后再试"
      ), !1;
    } finally {
      D(!1);
    }
  }
  R.useEffect(() => {
    r?.(P);
  });
  const [at, Ft] = R.useState(B1), [V, Ut] = R.useState(!0), [Bl, sl] = R.useState(""), [Qt, gl] = R.useState(""), [$l, Zt] = R.useState(
    /* @__PURE__ */ new Set()
  ), [x, H] = R.useState(
    null
  ), [K, dt] = R.useState(!1), [d, N] = R.useState(null), [Y, q] = R.useState(
    null
  ), [w, ct] = R.useState(null), [$, el] = R.useState(null), [xt, Yl] = R.useState(""), [Pa, ra] = R.useState(""), [Ia, Xe] = R.useState(""), [Sl, Pu] = R.useState(!1), [Qe, bl] = R.useState(""), [vi, tu] = R.useState(0), rl = R.useRef(null), Wl = R.useRef(
    null
  ), ye = R.useRef(!1), Zl = R.useRef(null), pl = R.useRef(null), Ze = R.useRef(null), ol = R.useRef(null);
  function ft(E) {
    Yl(E), window.setTimeout(() => Yl(""), 3600);
  }
  const Dt = async (E = !1) => {
    try {
      const X = await qt.state(), F = j ? {
        ...X,
        projects: X.projects.filter((it) => it.id === j.projectId),
        sessions: X.sessions.filter((it) => it.projectId === j.projectId),
        actions: X.actions.filter((it) => it.projectId === j.projectId),
        runs: X.runs.filter((it) => X.actions.some((rt) => rt.id === it.actionId && rt.projectId === j.projectId))
      } : X;
      Ft(F), sl(
        (it) => it && F.projects.some((rt) => rt.id === it) ? it : F.projects[0]?.id || ""
      ), gl(
        (it) => it && F.sessions.some((rt) => rt.id === it) ? it : j?.sessionId || F.sessions.find(
          (rt) => rt.projectId === (Bl || F.projects[0]?.id)
        )?.id || ""
      );
    } catch (X) {
      E || ft(X instanceof Error ? X.message : "无法连接服务");
    } finally {
      Ut(!1);
    }
  };
  R.useEffect(() => {
    Dt();
  }, []);
  const Pt = at.actions.some((E) => E.status === "running") || at.sessions.some(
    (E) => E.status === "recording" || ["queued", "running"].includes(E.analysisStatus || "")
  );
  R.useEffect(() => (Pt && !ol.current && (ol.current = window.setInterval(() => {
    Dt(!0);
  }, 1800)), !Pt && ol.current && (window.clearInterval(ol.current), ol.current = null), () => {
    ol.current && window.clearInterval(ol.current), ol.current = null;
  }), [Pt]);
  const ve = at.projects.find((E) => E.id === Bl) || at.projects[0], Ot = at.sessions.find(
    (E) => E.id === Qt
  ), Iu = at.actions.filter((E) => E.sessionId === Ot?.id).sort((E, X) => {
    const F = { running: 0, ready: 1, draft: 2, failed: 3, done: 4 };
    return F[E.status] - F[X.status] || X.createdAt.localeCompare(E.createdAt);
  }), oa = R.useMemo(
    () => new Map(
      at.projects.map((E) => [
        E.id,
        at.sessions.filter((X) => X.projectId === E.id)
      ])
    ),
    [at.projects, at.sessions]
  );
  R.useEffect(() => {
    if (!Sl || !Ze.current) return;
    const E = () => tu(
      Date.now() - new Date(Ze.current.startedAt).getTime()
    );
    E();
    const X = window.setInterval(E, 1e3);
    return () => window.clearInterval(X);
  }, [Sl]);
  function da(E) {
    Bt(!1), sl(E), gl(
      at.sessions.find((X) => X.projectId === E)?.id || ""
    ), Zt((X) => {
      const F = new Set(X);
      return F.delete(E), F;
    });
  }
  async function Ll() {
    if (ve)
      try {
        Sl && await we();
        const E = await qt.createSession(ve.id, "");
        gl(E.id), Bt(!1), await Dt(!0);
      } catch (E) {
        ft(E instanceof Error ? E.message : "创建会话失败");
      }
  }
  function Le(E, X) {
    const F = Date.now(), it = [], rt = new MediaRecorder(E, { audioBitsPerSecond: 64e3 });
    rl.current = rt, rt.ondataavailable = (dl) => {
      dl.data.size && it.push(dl.data);
    };
    let Lt = () => {
    };
    const ge = new Promise((dl) => {
      Lt = dl;
    });
    rt.saved = ge, rt.onstop = () => {
      const dl = Date.now(), Fl = new Blob(it, {
        type: rt.mimeType || "audio/webm"
      });
      if (!Fl.size) {
        Lt();
        return;
      }
      const il = qt.uploadAudio(X.id, Fl, {
        startedAt: new Date(F).toISOString(),
        endedAt: new Date(dl).toISOString(),
        durationMs: dl - F
      }).then(() => Dt(!0)).catch((Yt) => {
        const al = URL.createObjectURL(Fl);
        C.current.push({
          url: al,
          name: `sayso-${new Date(F).toISOString().replace(/:/g, "-")}.${Fl.type.includes("mp4") ? "m4a" : "webm"}`
        }), p([...C.current]), ft(`录音未上传：${Yt.message}。请下载备份。`);
      }).finally(() => {
        G.current.delete(il), Lt();
      });
      G.current.add(il);
    }, rt.start(1e3), pl.current = window.setTimeout(
      () => Ve(X),
      H1
    );
  }
  function Ve(E) {
    pl.current && window.clearTimeout(pl.current), pl.current = null;
    const X = Zl.current, F = rl.current;
    !X || !F || F.state !== "recording" || (F.stop(), Le(X, E));
  }
  async function mi(E) {
    if (Sl && await we(), !navigator.mediaDevices?.getUserMedia) {
      ft("当前浏览器不支持录音");
      return;
    }
    try {
      const X = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: !0, noiseSuppression: !0 }
      });
      Zl.current = X, await qt.updateSession(E.id, { status: "recording" }), Ze.current = E, ye.current = !1, Le(X, E);
      const F = window, it = F.SpeechRecognition || F.webkitSpeechRecognition;
      if (it) {
        const rt = new it();
        rt.continuous = !0, rt.interimResults = !0, rt.lang = "zh-CN", rt.onresult = (Lt) => {
          let ge = "";
          for (let dl = Lt.resultIndex; dl < Lt.results.length; dl += 1) {
            const Fl = Lt.results[dl], il = Fl[0].transcript.trim();
            Fl.isFinal && il ? qt.addUtterance(E.id, {
              text: il,
              atMs: Date.now() - new Date(E.startedAt).getTime()
            }).then(() => Dt(!0)).catch((Yt) => ft(Yt.message)) : ge += il;
          }
          Xe(ge);
        }, rt.onerror = (Lt) => {
          [
            "network",
            "service-not-allowed",
            "not-allowed",
            "audio-capture"
          ].includes(Lt.error) && (ye.current = !0, Xe(""), ft(
            Lt.error === "network" ? "实时转写服务网络不可用；录音仍会继续并正常保存" : "录音正常；当前浏览器暂时无法使用实时转写"
          ));
        }, rt.onend = () => {
          if (!ye.current && rl.current?.state === "recording")
            try {
              rt.start();
            } catch {
            }
        };
        try {
          rt.start();
        } catch {
          ye.current = !0, ft("录音已开始，实时转写不可用，可以打字补充。");
        }
        Wl.current = rt;
      } else ft("录音已开始；当前浏览器不提供实时转写");
      bl(E.id), Pu(!0), tu(Date.now() - new Date(E.startedAt).getTime()), await Dt(!0);
    } catch (X) {
      pl.current && window.clearTimeout(pl.current), rl.current?.state === "recording" && rl.current.stop(), await qt.updateSession(E.id, { status: "paused" }).catch(() => {
      }), Zl.current?.getTracks().forEach((F) => F.stop()), Zl.current = null, rl.current = null, ft(
        X instanceof Error ? `无法开始录音：${X.message}` : "无法开始录音"
      );
    }
  }
  async function we() {
    const E = Ze.current?.id || Qe, X = rl.current?.saved;
    ye.current = !0;
    try {
      Wl.current?.stop();
    } catch {
    }
    Wl.current = null, pl.current && window.clearTimeout(pl.current), pl.current = null, rl.current?.state !== "inactive" && rl.current?.stop(), Zl.current?.getTracks().forEach((F) => F.stop()), Zl.current = null, rl.current = null, Ze.current = null, Xe(""), Pu(!1), bl(""), await X, await Promise.allSettled([...G.current]), E && (await qt.updateSession(E, { status: "paused" }), await Dt(!0));
  }
  Z.current = we;
  async function me(E, X) {
    try {
      await E(), X && ft(X), await Dt(!0);
    } catch (F) {
      ft(F instanceof Error ? F.message : "操作失败");
    }
  }
  return V ? /* @__PURE__ */ c.jsxs("div", { className: "loading-screen", children: [
    /* @__PURE__ */ c.jsx("div", { className: "brand-mark", children: /* @__PURE__ */ c.jsx(_f, { size: 22 }) }),
    /* @__PURE__ */ c.jsx("p", { children: "正在打开 SaySo…" }),
    /* @__PURE__ */ c.jsx("button", { onClick: o, children: "返回工作台" })
  ] }) : /* @__PURE__ */ c.jsxs(
    "div",
    {
      className: `sayso-layout ${j ? "embedded" : ""} pane-${Wt} ${Et ? "tree-open" : ""}`,
      children: [
        /* @__PURE__ */ c.jsxs("header", { className: "remote-header", children: [
          /* @__PURE__ */ c.jsxs("button", { onClick: () => {
            P();
          }, disabled: _, children: [
            "← ",
            _ ? "正在保存…" : "工作台"
          ] }),
          /* @__PURE__ */ c.jsx("strong", { children: "SaySo" }),
          /* @__PURE__ */ c.jsx("span", { children: b }),
          /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "mobile-tree-button",
              onClick: () => Bt(!Et),
              children: "项目 / 会话"
            }
          ),
          j && /* @__PURE__ */ c.jsx("select", { "aria-label": "SaySo 历史讨论", value: Qt, onChange: (E) => {
            Sl || gl(E.target.value);
          }, disabled: Sl, children: at.sessions.map((E) => /* @__PURE__ */ c.jsx("option", { value: E.id, children: E.title }, E.id)) })
        ] }),
        M.length > 0 && /* @__PURE__ */ c.jsxs("div", { className: "audio-recovery", children: [
          "未上传录音，请下载保存：",
          M.map((E) => /* @__PURE__ */ c.jsx("a", { href: E.url, download: E.name, children: "下载录音备份" }, E.url))
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
                  onClick: () => dt(!0),
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
                  Ll();
                },
                disabled: !ve,
                children: [
                  /* @__PURE__ */ c.jsx(di, { size: 16 }),
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
                  onClick: () => H("new"),
                  "aria-label": "添加项目",
                  children: /* @__PURE__ */ c.jsx(di, { size: 14 })
                }
              )
            ] }),
            /* @__PURE__ */ c.jsx("div", { className: "project-tree", children: at.projects.map((E) => {
              const X = oa.get(E.id) || [], F = $l.has(E.id);
              return /* @__PURE__ */ c.jsxs("div", { className: "project-branch", children: [
                /* @__PURE__ */ c.jsxs(
                  "div",
                  {
                    className: `project-node ${ve?.id === E.id ? "selected" : ""}`,
                    children: [
                      /* @__PURE__ */ c.jsxs(
                        "button",
                        {
                          className: "project-select",
                          onClick: () => da(E.id),
                          children: [
                            /* @__PURE__ */ c.jsx(
                              "span",
                              {
                                className: "tree-chevron",
                                onClick: (it) => {
                                  it.stopPropagation(), Zt((rt) => {
                                    const Lt = new Set(rt);
                                    return Lt.has(E.id) ? Lt.delete(E.id) : Lt.add(E.id), Lt;
                                  });
                                },
                                children: F ? /* @__PURE__ */ c.jsx(Cf, { size: 13 }) : /* @__PURE__ */ c.jsx(hi, { size: 13 })
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
                          onClick: () => H(E),
                          "aria-label": `编辑项目 ${E.name}`,
                          children: /* @__PURE__ */ c.jsx(oi, { size: 13 })
                        }
                      )
                    ]
                  }
                ),
                !F && /* @__PURE__ */ c.jsxs("div", { className: "session-tree", children: [
                  X.map((it) => {
                    const rt = Sl && Qe === it.id;
                    return /* @__PURE__ */ c.jsxs(
                      "div",
                      {
                        className: `tree-session ${Ot?.id === it.id ? "active" : ""}`,
                        children: [
                          /* @__PURE__ */ c.jsxs(
                            "button",
                            {
                              className: "session-select",
                              onClick: () => {
                                sl(E.id), gl(it.id), Bt(!1);
                              },
                              children: [
                                /* @__PURE__ */ c.jsx(
                                  "span",
                                  {
                                    className: `tree-session-icon ${rt ? "recording" : ""}`,
                                    children: rt ? /* @__PURE__ */ c.jsx(Gd, { size: 12 }) : /* @__PURE__ */ c.jsx(Hf, { size: 12 })
                                  }
                                ),
                                /* @__PURE__ */ c.jsxs("span", { children: [
                                  /* @__PURE__ */ c.jsx("strong", { children: it.title }),
                                  /* @__PURE__ */ c.jsxs("small", { children: [
                                    qd(it.startedAt),
                                    " ·",
                                    " ",
                                    it.transcript.length,
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
                              onClick: () => N(it),
                              "aria-label": `重命名 ${it.title}`,
                              children: /* @__PURE__ */ c.jsx(oi, { size: 12 })
                            }
                          )
                        ]
                      },
                      it.id
                    );
                  }),
                  X.length === 0 && /* @__PURE__ */ c.jsx("div", { className: "tree-empty", children: "暂无会话" })
                ] })
              ] }, E.id);
            }) })
          ] }),
          /* @__PURE__ */ c.jsx("main", { className: "workspace", children: Ot ? /* @__PURE__ */ c.jsxs(c.Fragment, { children: [
            /* @__PURE__ */ c.jsxs("header", { className: "topbar", children: [
              /* @__PURE__ */ c.jsxs("div", { className: "title-stack", children: [
                /* @__PURE__ */ c.jsxs("div", { className: "eyebrow", children: [
                  /* @__PURE__ */ c.jsx("span", { children: ve?.name }),
                  /* @__PURE__ */ c.jsx("span", { children: "/" }),
                  /* @__PURE__ */ c.jsx("span", { children: qd(Ot.startedAt) })
                ] }),
                /* @__PURE__ */ c.jsxs("div", { className: "session-title-line", children: [
                  /* @__PURE__ */ c.jsx("h1", { children: Ot.title }),
                  /* @__PURE__ */ c.jsx(
                    "button",
                    {
                      className: "small-icon",
                      onClick: () => N(Ot),
                      "aria-label": "编辑会话名称",
                      children: /* @__PURE__ */ c.jsx(oi, { size: 14 })
                    }
                  )
                ] })
              ] }),
              /* @__PURE__ */ c.jsxs("div", { className: "topbar-actions", children: [
                (Ot.recordings?.length || Ot.recordingPath) && /* @__PURE__ */ c.jsxs(
                  "button",
                  {
                    className: "ghost-button",
                    onClick: () => el(Ot),
                    children: [
                      /* @__PURE__ */ c.jsx(b1, { size: 15 }),
                      " 录音片段"
                    ]
                  }
                ),
                Sl && Qe === Ot.id ? /* @__PURE__ */ c.jsxs("div", { className: "recording-control", children: [
                  /* @__PURE__ */ c.jsxs("span", { className: "recording-time", children: [
                    /* @__PURE__ */ c.jsx("i", {}),
                    " ",
                    Xd(vi)
                  ] }),
                  /* @__PURE__ */ c.jsxs(
                    "button",
                    {
                      className: "pause-button",
                      onClick: () => {
                        we().catch(
                          (E) => ft(E.message)
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
                      mi(Ot);
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
                  className: Wt === "discussion" ? "active" : "",
                  onClick: () => pt("discussion"),
                  children: "讨论记录"
                }
              ),
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: Wt === "actions" ? "active" : "",
                  onClick: () => pt("actions"),
                  children: [
                    "Action · ",
                    Iu.length
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "content-grid", children: [
              /* @__PURE__ */ c.jsx(
                w1,
                {
                  session: Ot,
                  interimText: Qe === Ot.id ? Ia : "",
                  isRecording: Sl && Qe === Ot.id,
                  search: Pa,
                  setSearch: ra
                }
              ),
              /* @__PURE__ */ c.jsx(
                K1,
                {
                  stewardMode: j?.ceo,
                  session: Ot,
                  actions: Iu,
                  runs: at.runs,
                  settings: at.settings,
                  onCreate: () => q("new"),
                  onExtract: () => me(
                    () => qt.extractActions(Ot.id),
                    "已开始重新分析讨论"
                  ),
                  onEdit: (E) => q(E),
                  onDelete: (E) => {
                    window.confirm(`删除“${E.title}”？`) && me(() => qt.deleteAction(E.id));
                  },
                  onReady: (E) => me(
                    () => qt.updateAction(E.id, {
                      status: E.status === "ready" ? "draft" : "ready"
                    })
                  ),
                  onDispatch: (E) => j?.ceo ? O(E) : me(
                    async () => {
                      const X = await qt.dispatchAction(E.id);
                      return X.webSessionId && await j?.onTaskCreated?.(X.webSessionId), X;
                    },
                    "已创建新的 Web 会话并排队执行"
                  ),
                  onViewRun: (E) => ct(E)
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs(
              "form",
              {
                className: "manual-discussion",
                onSubmit: async (E) => {
                  if (E.preventDefault(), !(!tt.trim() || yt)) {
                    Nt(!0);
                    try {
                      await qt.addUtterance(Ot.id, {
                        text: tt.trim(),
                        atMs: Date.now() - Date.parse(Ot.startedAt)
                      }), B(""), await Dt(!0);
                    } catch (X) {
                      ft(
                        X instanceof Error ? X.message : "记录失败"
                      );
                    } finally {
                      Nt(!1);
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
                      value: tt,
                      onChange: (E) => B(E.target.value)
                    }
                  ),
                  /* @__PURE__ */ c.jsx("button", { disabled: yt || !tt.trim(), children: "记录" }),
                  /* @__PURE__ */ c.jsx("small", { children: "录音保存在目标电脑。实时转写由浏览器提供，可能使用其在线识别服务。" })
                ]
              }
            )
          ] }) : /* @__PURE__ */ c.jsx(
            L1,
            {
              project: ve,
              onStart: () => {
                Ll();
              },
              onProject: () => H("new")
            }
          ) })
        ] }),
        U && /* @__PURE__ */ c.jsx(_1, { action: U, onClose: () => O(null), onSend: async (E) => {
          const X = await qt.dispatchAction(U.id, E);
          await j?.onTaskCreated?.(X.webSessionId || "", E.name || "下游管家"), await Dt(!0), ft("已加入派工队列，完成后回报 CEO"), O(null);
        } }),
        K && /* @__PURE__ */ c.jsx(
          $1,
          {
            settings: at.settings,
            onClose: () => dt(!1),
            onSave: async (E) => {
              try {
                await qt.updateSettings(E), dt(!1), await Dt(!0), ft("系统设置已保存");
              } catch (X) {
                ft(
                  X instanceof Error ? X.message : "保存设置失败"
                );
              }
            }
          }
        ),
        x && /* @__PURE__ */ c.jsx(
          P1,
          {
            project: x === "new" ? void 0 : x,
            onClose: () => H(null),
            onSave: async (E) => {
              try {
                const X = x === "new" ? await qt.createProject(E) : await qt.updateProject(x.id, E);
                sl(X.id), H(null), await Dt(!0);
              } catch (X) {
                ft(
                  X instanceof Error ? X.message : "保存项目失败"
                );
              }
            }
          }
        ),
        d && /* @__PURE__ */ c.jsx(
          I1,
          {
            session: d,
            onClose: () => N(null),
            onSave: async (E) => {
              try {
                await qt.updateSession(d.id, { title: E }), N(null), await Dt(!0);
              } catch (X) {
                ft(X instanceof Error ? X.message : "重命名失败");
              }
            }
          }
        ),
        Y && Ot && /* @__PURE__ */ c.jsx(
          ty,
          {
            action: Y === "new" ? void 0 : Y,
            onClose: () => q(null),
            onSave: async (E) => {
              await me(
                () => Y === "new" ? qt.createAction(Ot.id, E) : qt.updateAction(Y.id, E)
              ), q(null);
            }
          }
        ),
        w && /* @__PURE__ */ c.jsx(
          ly,
          {
            action: w,
            state: at,
            onClose: () => ct(null)
          }
        ),
        $ && /* @__PURE__ */ c.jsx(
          F1,
          {
            session: at.sessions.find((E) => E.id === $.id) || $,
            onClose: () => el(null)
          }
        ),
        xt && /* @__PURE__ */ c.jsxs("div", { className: "toast", children: [
          /* @__PURE__ */ c.jsx(Yd, { size: 16 }),
          " ",
          xt
        ] })
      ]
    }
  );
}
function L1({
  project: o,
  onStart: b,
  onProject: j
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
      "SaySo 会持续转写、分析，把确认的 Action 交给新的 Web 会话。"
    ] }),
    /* @__PURE__ */ c.jsxs("div", { className: "empty-actions", children: [
      /* @__PURE__ */ c.jsxs("button", { className: "primary-button", onClick: b, children: [
        /* @__PURE__ */ c.jsx(di, { size: 17 }),
        " 新建会话"
      ] }),
      !o && /* @__PURE__ */ c.jsxs("button", { className: "ghost-button", onClick: j, children: [
        /* @__PURE__ */ c.jsx(qf, { size: 16 }),
        " 添加项目"
      ] })
    ] })
  ] });
}
function V1({
  text: o,
  query: b
}) {
  return /* @__PURE__ */ c.jsx(c.Fragment, { children: C1(o, b).map(
    (j, r) => j.match ? /* @__PURE__ */ c.jsx(
      "mark",
      {
        className: "discussion-search-highlight",
        children: j.text
      },
      `${r}-${j.text}`
    ) : /* @__PURE__ */ c.jsx("span", { children: j.text }, `${r}-${j.text}`)
  ) });
}
function w1({
  session: o,
  interimText: b,
  isRecording: j,
  search: r,
  setSearch: _
}) {
  const D = R.useRef(null), U = R.useRef(/* @__PURE__ */ new Map()), O = Bf(r), M = R.useMemo(
    () => o.transcript.filter(
      (B) => U1(B.text, O)
    ),
    [o.transcript, O]
  ), p = R.useMemo(
    () => new Map(M.map((B, yt) => [B.id, yt])),
    [M]
  ), C = M.map((B) => B.id).join("|"), [G, Z] = R.useState(-1);
  R.useEffect(() => {
    Z(O && M.length ? 0 : -1);
  }, [O, o.id]), R.useEffect(() => {
    Z(
      (B) => M.length === 0 ? -1 : Math.min(Math.max(B, 0), M.length - 1)
    );
  }, [M.length]), R.useEffect(() => {
    if (G < 0) return;
    U.current.get(M[G]?.id)?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [G, C, O]), R.useEffect(() => {
    O || D.current?.scrollIntoView({ behavior: "smooth" });
  }, [o.transcript.length, b, O]);
  const tt = (B) => {
    M.length && Z(
      (yt) => yt < 0 ? B === 1 ? 0 : M.length - 1 : (yt + B + M.length) % M.length
    );
  };
  return /* @__PURE__ */ c.jsxs("section", { className: "transcript-panel", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "panel-header", children: [
      /* @__PURE__ */ c.jsxs("div", { children: [
        /* @__PURE__ */ c.jsx("h2", { children: "讨论记录" }),
        /* @__PURE__ */ c.jsxs("span", { children: [
          o.transcript.length,
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
                /* @__PURE__ */ c.jsx(x1, { size: 14 }),
                /* @__PURE__ */ c.jsx(
                  "input",
                  {
                    "aria-label": "搜索讨论",
                    value: r,
                    onChange: (B) => _(B.target.value),
                    onKeyDown: (B) => {
                      B.key === "Enter" && (B.preventDefault(), tt(B.shiftKey ? -1 : 1)), B.key === "Escape" && _("");
                    },
                    placeholder: "搜索讨论"
                  }
                )
              ] }),
              O && /* @__PURE__ */ c.jsxs("div", { className: "search-navigation", children: [
                /* @__PURE__ */ c.jsx("span", { children: M.length ? `${G + 1}/${M.length}` : "0/0" }),
                /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !M.length,
                    onClick: () => tt(-1),
                    "aria-label": "上一个搜索结果",
                    children: /* @__PURE__ */ c.jsx(d1, { size: 13 })
                  }
                ),
                /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !M.length,
                    onClick: () => tt(1),
                    "aria-label": "下一个搜索结果",
                    children: /* @__PURE__ */ c.jsx(hi, { size: 13 })
                  }
                )
              ] })
            ]
          }
        ),
        /* @__PURE__ */ c.jsxs("span", { className: `live-chip ${j ? "on" : ""}`, children: [
          /* @__PURE__ */ c.jsx("i", {}),
          " ",
          j ? "录音中" : "已暂停"
        ] })
      ] })
    ] }),
    /* @__PURE__ */ c.jsx("div", { className: "transcript-scroll", children: o.transcript.length === 0 && !b ? /* @__PURE__ */ c.jsxs("div", { className: "transcript-empty", children: [
      /* @__PURE__ */ c.jsxs("div", { className: `sound-wave ${j ? "active" : ""}`, children: [
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {}),
        /* @__PURE__ */ c.jsx("i", {})
      ] }),
      /* @__PURE__ */ c.jsx("h3", { children: j ? "正在听…" : "讨论已暂停" }),
      /* @__PURE__ */ c.jsx("p", { children: j ? "识别出的讨论会自动出现在这里。" : "点击右上角“继续录音”开始记录。" })
    ] }) : /* @__PURE__ */ c.jsxs("div", { className: "utterance-list", children: [
      o.transcript.map((B, yt) => {
        const Nt = Hd(B.createdAt), Wt = yt === 0 || Nt !== Hd(o.transcript[yt - 1].createdAt), pt = X1(B.createdAt), Et = p.get(B.id), Bt = Et !== void 0 && Et === G;
        return /* @__PURE__ */ c.jsxs(
          "article",
          {
            className: `utterance ${Et !== void 0 ? "discussion-search-match" : ""} ${Bt ? "current-search-match" : ""}`,
            ref: (P) => {
              P ? U.current.set(B.id, P) : U.current.delete(B.id);
            },
            children: [
              /* @__PURE__ */ c.jsx(
                "time",
                {
                  className: Wt ? "minute-mark" : void 0,
                  dateTime: B.createdAt,
                  "aria-label": Wt ? `系统时间 ${pt}` : void 0,
                  "aria-hidden": !Wt || void 0,
                  children: Wt ? pt : null
                }
              ),
              /* @__PURE__ */ c.jsx("p", { children: /* @__PURE__ */ c.jsx(
                V1,
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
      b && /* @__PURE__ */ c.jsxs("article", { className: "utterance interim", children: [
        /* @__PURE__ */ c.jsxs("time", { children: [
          /* @__PURE__ */ c.jsx(Gd, { size: 10 }),
          " LIVE"
        ] }),
        /* @__PURE__ */ c.jsxs("p", { children: [
          b,
          /* @__PURE__ */ c.jsx("span", { className: "typing-cursor" })
        ] })
      ] }),
      /* @__PURE__ */ c.jsx("div", { ref: D })
    ] }) })
  ] });
}
function K1({
  stewardMode: o,
  session: b,
  actions: j,
  runs: r,
  settings: _,
  onCreate: D,
  onExtract: U,
  onEdit: O,
  onDelete: M,
  onReady: p,
  onDispatch: C,
  onViewRun: G
}) {
  const Z = ["queued", "running"].includes(
    b.analysisStatus || ""
  ), [tt, B] = R.useState(!0), [yt, Nt] = R.useState(!1), pt = b.analysisTraces?.find(
    (V) => V.id === b.analysisActiveTraceId
  ) || b.analysisTraces?.[0], Et = R.useRef(void 0);
  R.useEffect(() => {
    Z && pt?.id && Et.current !== pt.id && (Et.current = pt.id, B(!0));
  }, [Z, pt?.id]);
  const Bt = [
    "读取最近 5 分钟的原始讨论",
    `对照 ${j.filter((V) => !["running", "done"].includes(V.status)).length} 条已有需求`,
    `${_.analysisModel === "default" ? "Codex" : _.analysisModel} 正在判断补充与修正`
  ], P = pt && pt.events.length >= 3 ? pt.events.slice(-3) : Bt, at = Zd(
    P,
    Z && tt
  ), Ft = [...at].reverse().find(Boolean);
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
          onClick: D,
          "aria-label": "添加 Action",
          children: /* @__PURE__ */ c.jsx(di, { size: 18 })
        }
      )
    ] }),
    /* @__PURE__ */ c.jsxs(
      "div",
      {
        className: `analysis-strip ${b.analysisStatus === "failed" ? "failed" : ""} ${tt ? "expanded" : ""}`,
        children: [
          Z ? /* @__PURE__ */ c.jsx(he, { size: 13, className: "spin" }) : /* @__PURE__ */ c.jsx(Ud, { size: 13 }),
          /* @__PURE__ */ c.jsxs(
            "button",
            {
              className: "analysis-status-copy",
              onClick: () => Z && B((V) => !V),
              disabled: !Z,
              children: [
                /* @__PURE__ */ c.jsx("span", { children: Z ? tt && Ft || "分析中" : b.analysisStatus === "failed" ? b.analysisError || "分析失败，点击重试" : b.status === "recording" ? "每 2 分钟分析最近 5 分钟讨论" : `${j.length} 条 Action` }),
                Z && (tt ? /* @__PURE__ */ c.jsx(hi, { size: 12 }) : /* @__PURE__ */ c.jsx(Cf, { size: 12 }))
              ]
            }
          ),
          pt && /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "analysis-details-button",
              onClick: () => Nt(!0),
              children: "详情"
            }
          ),
          /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "analysis-refresh",
              onClick: U,
              title: "重新分析讨论",
              children: /* @__PURE__ */ c.jsx(p1, { size: 13 })
            }
          ),
          Z && tt && /* @__PURE__ */ c.jsx("div", { className: "analysis-detail-lines", children: P.map((V, Ut) => /* @__PURE__ */ c.jsxs(
            "div",
            {
              className: at[Ut] === V ? "complete" : at[Ut] ? "typing" : "",
              children: [
                /* @__PURE__ */ c.jsx("span", { children: Ut + 1 }),
                /* @__PURE__ */ c.jsx("p", { children: at[Ut] || " " })
              ]
            },
            `${Ut}-${V}`
          )) })
        ]
      }
    ),
    /* @__PURE__ */ c.jsx("div", { className: "action-list", children: j.length === 0 ? /* @__PURE__ */ c.jsxs("div", { className: "action-empty-compact", children: [
      /* @__PURE__ */ c.jsx(Ud, { size: 17 }),
      /* @__PURE__ */ c.jsx("span", { children: Z ? "正在整理讨论…" : "还没有明确的 Action" })
    ] }) : j.map((V) => {
      const Ut = r.find((Bl) => Bl.id === V.runId);
      return /* @__PURE__ */ c.jsxs(
        "article",
        {
          className: `action-card card-${V.status}`,
          children: [
            /* @__PURE__ */ c.jsxs("div", { className: "action-card-top", children: [
              /* @__PURE__ */ c.jsx(Qd, { status: V.status }),
              /* @__PURE__ */ c.jsx("span", { className: `priority priority-${V.priority}`, children: V.priority === "high" ? "高优先级" : V.priority === "low" ? "低优先级" : "普通" }),
              /* @__PURE__ */ c.jsxs("div", { className: "card-tools", children: [
                !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsx("button", { onClick: () => O(V), "aria-label": "编辑", children: /* @__PURE__ */ c.jsx(oi, { size: 14 }) }),
                !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsx(
                  "button",
                  {
                    onClick: () => M(V),
                    "aria-label": "删除",
                    children: /* @__PURE__ */ c.jsx(E1, { size: 14 })
                  }
                )
              ] })
            ] }),
            /* @__PURE__ */ c.jsx(
              J1,
              {
                action: V,
                trace: pt,
                active: Z && !["running", "done"].includes(V.status)
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
                onClick: () => G(V),
                children: [
                  /* @__PURE__ */ c.jsxs("span", { children: [
                    /* @__PURE__ */ c.jsx(he, { size: 15 }),
                    " Web 会话正在项目中工作"
                  ] }),
                  /* @__PURE__ */ c.jsx("small", { children: "查看输出" })
                ]
              }
            ),
            V.status === "done" && /* @__PURE__ */ c.jsxs(
              "button",
              {
                className: "run-complete",
                onClick: () => G(V),
                children: [
                  /* @__PURE__ */ c.jsxs("span", { children: [
                    /* @__PURE__ */ c.jsx(Yd, { size: 15 }),
                    " Web 会话已完成"
                  ] }),
                  /* @__PURE__ */ c.jsx("small", { children: "查看结果" })
                ]
              }
            ),
            V.status === "failed" && Ut?.error && /* @__PURE__ */ c.jsx("p", { className: "run-error", children: Ut.error }),
            !["running", "done"].includes(V.status) && /* @__PURE__ */ c.jsxs("div", { className: "action-card-footer", children: [
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: `ready-toggle ${V.status === "ready" ? "checked" : ""}`,
                  onClick: () => p(V),
                  children: [
                    /* @__PURE__ */ c.jsx("span", { children: V.status === "ready" && /* @__PURE__ */ c.jsx(Fu, { size: 11 }) }),
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
                  onClick: () => C(V),
                  children: [
                    /* @__PURE__ */ c.jsx(j1, { size: 14 }),
                    " ",
                    o ? "派给管家" : V.status === "failed" ? "重新创建 Web 会话" : "创建 Web 会话"
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
      /* @__PURE__ */ c.jsx("span", { children: "新 Web 会话使用当前项目的执行器与本机权限" })
    ] }),
    yt && pt && /* @__PURE__ */ c.jsx(
      k1,
      {
        trace: pt,
        onClose: () => Nt(!1)
      }
    )
  ] });
}
function J1({
  action: o,
  trace: b,
  active: j
}) {
  const [r, _] = R.useState(!0), D = R.useRef(void 0);
  R.useEffect(() => {
    j && b?.id && D.current !== b.id && (D.current = b.id, _(!0));
  }, [j, b?.id]);
  const U = [
    `输入：核对${b?.contextLabel || "最近 5 分钟"}的 ${b?.utteranceCount || 0} 条讨论`,
    `判断“${o.title}”是否需要补充或修正`
  ], O = Zd(U, j && r);
  return j ? /* @__PURE__ */ c.jsxs("div", { className: `requirement-analysis ${r ? "expanded" : ""}`, children: [
    /* @__PURE__ */ c.jsxs(
      "button",
      {
        className: "requirement-analysis-toggle",
        onClick: () => _((M) => !M),
        children: [
          /* @__PURE__ */ c.jsx(he, { size: 11, className: "spin" }),
          /* @__PURE__ */ c.jsx("span", { children: r ? "需求分析中" : "分析中" }),
          r ? /* @__PURE__ */ c.jsx(hi, { size: 11 }) : /* @__PURE__ */ c.jsx(Cf, { size: 11 })
        ]
      }
    ),
    r && /* @__PURE__ */ c.jsx("div", { className: "requirement-analysis-lines", children: U.map((M, p) => /* @__PURE__ */ c.jsx(
      "p",
      {
        className: O[p] === M ? "complete" : O[p] ? "typing" : "",
        children: O[p] || " "
      },
      M
    )) })
  ] }) : null;
}
function k1({
  trace: o,
  onClose: b
}) {
  const j = R.useRef(null);
  R.useEffect(() => {
    j.current?.scrollIntoView({
      behavior: "smooth",
      block: "nearest"
    });
  }, [o.events.length, o.output]);
  const r = {
    queued: "排队中",
    running: "分析中",
    done: "已完成",
    failed: "失败"
  }[o.status];
  return /* @__PURE__ */ c.jsxs(
    sa,
    {
      wide: !0,
      title: "分析详情",
      description: `${o.model} · ${r} · ${o.utteranceCount} 条原始讨论`,
      onClose: b,
      children: [
        /* @__PURE__ */ c.jsxs("div", { className: "analysis-trace-scroll", children: [
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsxs("div", { className: "analysis-trace-heading", children: [
              /* @__PURE__ */ c.jsx("h3", { children: "实时过程" }),
              /* @__PURE__ */ c.jsxs("span", { className: `trace-status trace-${o.status}`, children: [
                o.status === "running" && /* @__PURE__ */ c.jsx(he, { size: 11, className: "spin" }),
                r
              ] })
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "analysis-live-log", children: [
              o.events.map((_, D) => /* @__PURE__ */ c.jsxs("div", { children: [
                /* @__PURE__ */ c.jsx("span", { children: String(D + 1).padStart(2, "0") }),
                /* @__PURE__ */ c.jsx("p", { children: _ })
              ] }, `${D}-${_}`)),
              /* @__PURE__ */ c.jsx("div", { ref: j })
            ] })
          ] }),
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsx("h3", { children: "原始输入" }),
            /* @__PURE__ */ c.jsx("pre", { children: o.input || "暂无输入" })
          ] }),
          /* @__PURE__ */ c.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ c.jsx("h3", { children: "最终输出" }),
            /* @__PURE__ */ c.jsx("pre", { children: o.output || (o.status === "running" ? "等待模型返回最终结构化输出…" : o.error || "本轮没有输出") })
          ] })
        ] }),
        /* @__PURE__ */ c.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: b, children: "关闭" }) })
      ]
    }
  );
}
function $1({
  settings: o,
  onClose: b,
  onSave: j
}) {
  const [r, _] = R.useState(o), [D, U] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    sa,
    {
      title: "系统设置",
      description: "配置讨论如何被分析，以及 Action 由谁执行。",
      onClose: b,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), U(!0), j(r).finally(() => U(!1));
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
                  value: r.analysisProvider,
                  onChange: (O) => _({
                    ...r,
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
                  disabled: r.analysisProvider === "local",
                  value: r.analysisModel,
                  onChange: (O) => _({ ...r, analysisModel: O.target.value }),
                  children: Y1.map((O) => /* @__PURE__ */ c.jsx("option", { value: O.value, children: O.label }, O.value))
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
                  value: r.executionModel,
                  onChange: (O) => _({ ...r, executionModel: O.target.value }),
                  children: G1.map((O) => /* @__PURE__ */ c.jsx("option", { value: O.value, children: O.label }, O.value))
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
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: D, children: [
                D ? /* @__PURE__ */ c.jsx(he, { size: 15 }) : /* @__PURE__ */ c.jsx(Fu, { size: 15 }),
                " 保存设置"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function W1({
  recording: o,
  index: b
}) {
  const [j, r] = R.useState(""), [_, D] = R.useState(""), [U, O] = R.useState(!1), M = R.useRef(null);
  R.useEffect(
    () => () => {
      M.current?.abort();
    },
    []
  ), R.useEffect(
    () => () => {
      j && URL.revokeObjectURL(j);
    },
    [j]
  );
  const p = async () => {
    O(!0), D("");
    const B = new AbortController();
    M.current = B;
    try {
      r(await R1(o.id, B.signal));
    } catch (yt) {
      B.signal.aborted || D(yt instanceof Error ? yt.message : "无法读取录音");
    } finally {
      B.signal.aborted || O(!1);
    }
  }, [C, G] = R.useState(), Z = o.durationMs ?? C, tt = (B) => {
    if (Number.isFinite(B.duration) && B.duration > 0) {
      G(B.duration * 1e3), B.dataset.durationProbe === "active" && (B.currentTime = 0, delete B.dataset.durationProbe);
      return;
    }
    o.durationMs === void 0 && B.dataset.durationProbe !== "active" && (B.dataset.durationProbe = "active", B.currentTime = Number.MAX_SAFE_INTEGER);
  };
  return /* @__PURE__ */ c.jsxs("article", { className: "recording-row", children: [
    /* @__PURE__ */ c.jsxs("div", { className: "recording-index", children: [
      /* @__PURE__ */ c.jsx(Hf, { size: 15 }),
      /* @__PURE__ */ c.jsx("span", { children: String(b + 1).padStart(2, "0") })
    ] }),
    /* @__PURE__ */ c.jsxs("div", { className: "recording-info", children: [
      /* @__PURE__ */ c.jsxs("div", { children: [
        /* @__PURE__ */ c.jsxs("strong", { children: [
          "录音片段 ",
          b + 1
        ] }),
        /* @__PURE__ */ c.jsx("span", { children: new Date(o.startedAt).toLocaleString("zh-CN", {
          month: "numeric",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit"
        }) })
      ] }),
      /* @__PURE__ */ c.jsxs("small", { children: [
        Z !== void 0 ? Xd(Z) : "读取时长中",
        " ·",
        " ",
        Q1(o.sizeBytes)
      ] }),
      !j && /* @__PURE__ */ c.jsx(
        "button",
        {
          className: "ghost-button",
          disabled: U,
          onClick: () => {
            p();
          },
          children: U ? "正在从设备读取…" : "加载录音"
        }
      ),
      _ && /* @__PURE__ */ c.jsx("p", { role: "alert", children: _ }),
      j && /* @__PURE__ */ c.jsx(
        "audio",
        {
          controls: !0,
          preload: "metadata",
          src: j,
          onLoadedMetadata: (B) => tt(B.currentTarget),
          onDurationChange: (B) => tt(B.currentTarget),
          onSeeked: (B) => tt(B.currentTarget),
          onPlay: (B) => {
            B.currentTarget.getRootNode().querySelectorAll("audio").forEach((yt) => {
              yt !== B.currentTarget && yt.pause();
            });
          }
        }
      )
    ] })
  ] });
}
function F1({
  session: o,
  onClose: b
}) {
  const j = [...o.recordings || []].sort(
    (r, _) => r.startedAt.localeCompare(_.startedAt)
  );
  return /* @__PURE__ */ c.jsxs(
    sa,
    {
      wide: !0,
      title: "录音片段",
      description: `${o.title} · 共 ${j.length} 段，每段最长 10 分钟`,
      onClose: b,
      children: [
        /* @__PURE__ */ c.jsx("div", { className: "recordings-list", children: j.length ? j.map((r, _) => /* @__PURE__ */ c.jsx(
          W1,
          {
            recording: r,
            index: _
          },
          r.id
        )) : /* @__PURE__ */ c.jsxs("div", { className: "recordings-empty", children: [
          /* @__PURE__ */ c.jsx(Hf, { size: 20 }),
          "暂时没有录音"
        ] }) }),
        /* @__PURE__ */ c.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: b, children: "关闭" }) })
      ]
    }
  );
}
function P1({
  project: o,
  onClose: b,
  onSave: j
}) {
  const [r, _] = R.useState(o?.name || ""), [D, U] = R.useState(o?.rootPath || ""), [O, M] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    sa,
    {
      title: o ? "编辑项目" : "添加项目",
      description: "Action 将通过这个项目的执行器创建新的 Web 会话。",
      onClose: b,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (p) => {
            p.preventDefault(), M(!0), j({ name: r, rootPath: D }).finally(() => M(!1));
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "项目名称" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                value: r,
                onChange: (p) => _(p.target.value),
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
                  onChange: (p) => U(p.target.value),
                  placeholder: "C:\\\\Code\\\\my-project"
                }
              )
            ] }),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ c.jsxs(
                "button",
                {
                  className: "primary-button",
                  disabled: O || !r.trim() || !D.trim(),
                  children: [
                    O ? /* @__PURE__ */ c.jsx(he, { size: 15 }) : /* @__PURE__ */ c.jsx(Fu, { size: 15 }),
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
function I1({
  session: o,
  onClose: b,
  onSave: j
}) {
  const [r, _] = R.useState(o.title), [D, U] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    sa,
    {
      title: "重命名会话",
      description: "日期时间名称也可以随时改掉。",
      onClose: b,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), U(!0), j(r).finally(() => U(!1));
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "会话名称" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                autoFocus: !0,
                value: r,
                onChange: (O) => _(O.target.value)
              }
            ),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: D || !r.trim(), children: [
                D ? /* @__PURE__ */ c.jsx(he, { size: 15 }) : /* @__PURE__ */ c.jsx(Fu, { size: 15 }),
                " 保存"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function ty({
  action: o,
  onClose: b,
  onSave: j
}) {
  const [r, _] = R.useState(o?.title || ""), [D, U] = R.useState(o?.detail || ""), [O, M] = R.useState(
    o?.priority || "medium"
  ), [p, C] = R.useState(!1);
  return /* @__PURE__ */ c.jsx(
    sa,
    {
      title: o ? "完善 Action" : "添加 Action",
      description: "说明目标和验收结果，新 Web 会话将据此执行。",
      onClose: b,
      children: /* @__PURE__ */ c.jsxs(
        "form",
        {
          onSubmit: (G) => {
            G.preventDefault(), C(!0), j({ title: r, detail: D, priority: O }).finally(
              () => C(!1)
            );
          },
          children: [
            /* @__PURE__ */ c.jsx("label", { className: "field-label", children: "要做什么" }),
            /* @__PURE__ */ c.jsx(
              "input",
              {
                className: "text-field",
                value: r,
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
                onClick: () => M(G),
                children: [
                  /* @__PURE__ */ c.jsx(h1, { size: 9, fill: "currentColor" }),
                  " ",
                  G === "high" ? "高" : G === "medium" ? "普通" : "低"
                ]
              },
              G
            )) }),
            /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ c.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ c.jsxs("button", { className: "primary-button", disabled: p || !r.trim(), children: [
                p ? /* @__PURE__ */ c.jsx(he, { size: 15 }) : /* @__PURE__ */ c.jsx(Fu, { size: 15 }),
                " 保存 Action"
              ] })
            ] })
          ]
        }
      )
    }
  );
}
function ly({
  action: o,
  state: b,
  onClose: j
}) {
  const [r, _] = R.useState("");
  o = b.actions.find((U) => U.id === o.id) || o;
  const D = b.runs.find((U) => U.id === o.runId);
  return /* @__PURE__ */ c.jsxs(
    sa,
    {
      wide: !0,
      title: o.status === "running" ? "Web 会话正在工作" : o.status === "done" ? "Web 会话已完成" : "Web 会话执行结果",
      description: o.title,
      onClose: j,
      children: [
        /* @__PURE__ */ c.jsxs("div", { className: "run-meta", children: [
          /* @__PURE__ */ c.jsx(Qd, { status: o.status }),
          D && /* @__PURE__ */ c.jsxs("span", { children: [
            /* @__PURE__ */ c.jsx(y1, { size: 13 }),
            " ",
            new Date(D.startedAt).toLocaleString("zh-CN")
          ] })
        ] }),
        /* @__PURE__ */ c.jsx("pre", { className: "run-output", children: D?.output || (D?.status === "running" ? "正在创建 Web 会话…" : D?.error || "暂无输出") }),
        D?.error && /* @__PURE__ */ c.jsx("p", { role: "alert", children: D.error }),
        r && /* @__PURE__ */ c.jsx("p", { role: "alert", children: r }),
        /* @__PURE__ */ c.jsxs("div", { className: "modal-actions", children: [
          D?.status === "running" && /* @__PURE__ */ c.jsx(
            "button",
            {
              className: "ghost-button",
              onClick: () => {
                qt.cancelRun(D.id).catch((U) => _(U.message));
              },
              children: "停止执行"
            }
          ),
          /* @__PURE__ */ c.jsx("button", { className: "primary-button", onClick: j, children: "关闭" })
        ] })
      ]
    }
  );
}
function ey(o, b, j, r, _) {
  M1(b);
  const D = o.shadowRoot || o.attachShadow({ mode: "open" });
  D.replaceChildren();
  const U = document.createElement("link");
  U.rel = "stylesheet", U.href = new URL(
    /* @vite-ignore */
    "./sayso.css",
    import.meta.url
  ).href;
  const O = document.createElement("div");
  O.id = "root", D.append(U, O);
  const M = f1.createRoot(O);
  let p;
  o.toggleAttribute("data-embedded", !!_), M.render(/* @__PURE__ */ c.jsx(Z1, { onClose: j, deviceLabel: r, scope: _, registerClose: (G) => {
    p = G;
  } }));
  const C = () => M.unmount();
  return C.close = async () => p ? p() : !1, C;
}
export {
  ey as mount
};
