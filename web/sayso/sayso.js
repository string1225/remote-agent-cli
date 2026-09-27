var Ef = { exports: {} }, Ju = {};
var jd;
function I0() {
  if (jd) return Ju;
  jd = 1;
  var o = /* @__PURE__ */ Symbol.for("react.transitional.element"), b = /* @__PURE__ */ Symbol.for("react.fragment");
  function z(r, U, D) {
    var _ = null;
    if (D !== void 0 && (_ = "" + D), U.key !== void 0 && (_ = "" + U.key), "key" in U) {
      D = {};
      for (var O in U)
        O !== "key" && (D[O] = U[O]);
    } else D = U;
    return U = D.ref, {
      $$typeof: o,
      type: r,
      key: _,
      ref: U !== void 0 ? U : null,
      props: D
    };
  }
  return Ju.Fragment = b, Ju.jsx = z, Ju.jsxs = z, Ju;
}
var Td;
function t1() {
  return Td || (Td = 1, Ef.exports = I0()), Ef.exports;
}
var i = t1(), zf = { exports: {} }, ku = {}, Mf = { exports: {} }, Nf = {};
var Ad;
function l1() {
  return Ad || (Ad = 1, (function(o) {
    function b(p, C) {
      var V = p.length;
      p.push(C);
      t: for (; 0 < V; ) {
        var st = V - 1 >>> 1, d = p[st];
        if (0 < U(d, C))
          p[st] = C, p[V] = d, V = st;
        else break t;
      }
    }
    function z(p) {
      return p.length === 0 ? null : p[0];
    }
    function r(p) {
      if (p.length === 0) return null;
      var C = p[0], V = p.pop();
      if (V !== C) {
        p[0] = V;
        t: for (var st = 0, d = p.length, M = d >>> 1; st < M; ) {
          var Y = 2 * (st + 1) - 1, H = p[Y], w = Y + 1, ut = p[w];
          if (0 > U(H, V))
            w < d && 0 > U(ut, H) ? (p[st] = ut, p[w] = V, st = w) : (p[st] = H, p[Y] = V, st = Y);
          else if (w < d && 0 > U(ut, V))
            p[st] = ut, p[w] = V, st = w;
          else break t;
        }
      }
      return C;
    }
    function U(p, C) {
      var V = p.sortIndex - C.sortIndex;
      return V !== 0 ? V : p.id - C.id;
    }
    if (o.unstable_now = void 0, typeof performance == "object" && typeof performance.now == "function") {
      var D = performance;
      o.unstable_now = function() {
        return D.now();
      };
    } else {
      var _ = Date, O = _.now();
      o.unstable_now = function() {
        return _.now() - O;
      };
    }
    var N = [], x = [], B = 1, X = null, W = 3, it = !1, q = !1, ot = !1, Dt = !1, xt = typeof setTimeout == "function" ? setTimeout : null, ft = typeof clearTimeout == "function" ? clearTimeout : null, Et = typeof setImmediate < "u" ? setImmediate : null;
    function Lt(p) {
      for (var C = z(x); C !== null; ) {
        if (C.callback === null) r(x);
        else if (C.startTime <= p)
          r(x), C.sortIndex = C.expirationTime, b(N, C);
        else break;
        C = z(x);
      }
    }
    function P(p) {
      if (ot = !1, Lt(p), !q)
        if (z(N) !== null)
          q = !0, Vt || (Vt = !0, Ot());
        else {
          var C = z(x);
          C !== null && Bt(P, C.startTime - p);
        }
    }
    var Vt = !1, Z = -1, zt = 5, $t = -1;
    function ge() {
      return Dt ? !0 : !(o.unstable_now() - $t < zt);
    }
    function Al() {
      if (Dt = !1, Vt) {
        var p = o.unstable_now();
        $t = p;
        var C = !0;
        try {
          t: {
            q = !1, ot && (ot = !1, ft(Z), Z = -1), it = !0;
            var V = W;
            try {
              l: {
                for (Lt(p), X = z(N); X !== null && !(X.expirationTime > p && ge()); ) {
                  var st = X.callback;
                  if (typeof st == "function") {
                    X.callback = null, W = X.priorityLevel;
                    var d = st(
                      X.expirationTime <= p
                    );
                    if (p = o.unstable_now(), typeof d == "function") {
                      X.callback = d, Lt(p), C = !0;
                      break l;
                    }
                    X === z(N) && r(N), Lt(p);
                  } else r(N);
                  X = z(N);
                }
                if (X !== null) C = !0;
                else {
                  var M = z(x);
                  M !== null && Bt(
                    P,
                    M.startTime - p
                  ), C = !1;
                }
              }
              break t;
            } finally {
              X = null, W = V, it = !1;
            }
            C = void 0;
          }
        } finally {
          C ? Ot() : Vt = !1;
        }
      }
    }
    var Ot;
    if (typeof Et == "function")
      Ot = function() {
        Et(Al);
      };
    else if (typeof MessageChannel < "u") {
      var vl = new MessageChannel(), Il = vl.port2;
      vl.port1.onmessage = Al, Ot = function() {
        Il.postMessage(null);
      };
    } else
      Ot = function() {
        xt(Al, 0);
      };
    function Bt(p, C) {
      Z = xt(function() {
        p(o.unstable_now());
      }, C);
    }
    o.unstable_IdlePriority = 5, o.unstable_ImmediatePriority = 1, o.unstable_LowPriority = 4, o.unstable_NormalPriority = 3, o.unstable_Profiling = null, o.unstable_UserBlockingPriority = 2, o.unstable_cancelCallback = function(p) {
      p.callback = null;
    }, o.unstable_forceFrameRate = function(p) {
      0 > p || 125 < p ? console.error(
        "forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"
      ) : zt = 0 < p ? Math.floor(1e3 / p) : 5;
    }, o.unstable_getCurrentPriorityLevel = function() {
      return W;
    }, o.unstable_next = function(p) {
      switch (W) {
        case 1:
        case 2:
        case 3:
          var C = 3;
          break;
        default:
          C = W;
      }
      var V = W;
      W = C;
      try {
        return p();
      } finally {
        W = V;
      }
    }, o.unstable_requestPaint = function() {
      Dt = !0;
    }, o.unstable_runWithPriority = function(p, C) {
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
      var V = W;
      W = p;
      try {
        return C();
      } finally {
        W = V;
      }
    }, o.unstable_scheduleCallback = function(p, C, V) {
      var st = o.unstable_now();
      switch (typeof V == "object" && V !== null ? (V = V.delay, V = typeof V == "number" && 0 < V ? st + V : st) : V = st, p) {
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
      return d = V + d, p = {
        id: B++,
        callback: C,
        priorityLevel: p,
        startTime: V,
        expirationTime: d,
        sortIndex: -1
      }, V > st ? (p.sortIndex = V, b(x, p), z(N) === null && p === z(x) && (ot ? (ft(Z), Z = -1) : ot = !0, Bt(P, V - st))) : (p.sortIndex = d, b(N, p), q || it || (q = !0, Vt || (Vt = !0, Ot()))), p;
    }, o.unstable_shouldYield = ge, o.unstable_wrapCallback = function(p) {
      var C = W;
      return function() {
        var V = W;
        W = C;
        try {
          return p.apply(this, arguments);
        } finally {
          W = V;
        }
      };
    };
  })(Nf)), Nf;
}
var Ed;
function e1() {
  return Ed || (Ed = 1, Mf.exports = l1()), Mf.exports;
}
var Df = { exports: {} }, I = {};
var zd;
function a1() {
  if (zd) return I;
  zd = 1;
  var o = /* @__PURE__ */ Symbol.for("react.transitional.element"), b = /* @__PURE__ */ Symbol.for("react.portal"), z = /* @__PURE__ */ Symbol.for("react.fragment"), r = /* @__PURE__ */ Symbol.for("react.strict_mode"), U = /* @__PURE__ */ Symbol.for("react.profiler"), D = /* @__PURE__ */ Symbol.for("react.consumer"), _ = /* @__PURE__ */ Symbol.for("react.context"), O = /* @__PURE__ */ Symbol.for("react.forward_ref"), N = /* @__PURE__ */ Symbol.for("react.suspense"), x = /* @__PURE__ */ Symbol.for("react.memo"), B = /* @__PURE__ */ Symbol.for("react.lazy"), X = Symbol.iterator;
  function W(d) {
    return d === null || typeof d != "object" ? null : (d = X && d[X] || d["@@iterator"], typeof d == "function" ? d : null);
  }
  var it = {
    isMounted: function() {
      return !1;
    },
    enqueueForceUpdate: function() {
    },
    enqueueReplaceState: function() {
    },
    enqueueSetState: function() {
    }
  }, q = Object.assign, ot = {};
  function Dt(d, M, Y) {
    this.props = d, this.context = M, this.refs = ot, this.updater = Y || it;
  }
  Dt.prototype.isReactComponent = {}, Dt.prototype.setState = function(d, M) {
    if (typeof d != "object" && typeof d != "function" && d != null)
      throw Error(
        "takes an object of state variables to update or a function which returns an object of state variables."
      );
    this.updater.enqueueSetState(this, d, M, "setState");
  }, Dt.prototype.forceUpdate = function(d) {
    this.updater.enqueueForceUpdate(this, d, "forceUpdate");
  };
  function xt() {
  }
  xt.prototype = Dt.prototype;
  function ft(d, M, Y) {
    this.props = d, this.context = M, this.refs = ot, this.updater = Y || it;
  }
  var Et = ft.prototype = new xt();
  Et.constructor = ft, q(Et, Dt.prototype), Et.isPureReactComponent = !0;
  var Lt = Array.isArray, P = { H: null, A: null, T: null, S: null, V: null }, Vt = Object.prototype.hasOwnProperty;
  function Z(d, M, Y, H, w, ut) {
    return Y = ut.ref, {
      $$typeof: o,
      type: d,
      key: M,
      ref: Y !== void 0 ? Y : null,
      props: ut
    };
  }
  function zt(d, M) {
    return Z(
      d.type,
      M,
      void 0,
      void 0,
      void 0,
      d.props
    );
  }
  function $t(d) {
    return typeof d == "object" && d !== null && d.$$typeof === o;
  }
  function ge(d) {
    var M = { "=": "=0", ":": "=2" };
    return "$" + d.replace(/[=:]/g, function(Y) {
      return M[Y];
    });
  }
  var Al = /\/+/g;
  function Ot(d, M) {
    return typeof d == "object" && d !== null && d.key != null ? ge("" + d.key) : M.toString(36);
  }
  function vl() {
  }
  function Il(d) {
    switch (d.status) {
      case "fulfilled":
        return d.value;
      case "rejected":
        throw d.reason;
      default:
        switch (typeof d.status == "string" ? d.then(vl, vl) : (d.status = "pending", d.then(
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
  function Bt(d, M, Y, H, w) {
    var ut = typeof d;
    (ut === "undefined" || ut === "boolean") && (d = null);
    var $ = !1;
    if (d === null) $ = !0;
    else
      switch (ut) {
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
            case B:
              return $ = d._init, Bt(
                $(d._payload),
                M,
                Y,
                H,
                w
              );
          }
      }
    if ($)
      return w = w(d), $ = H === "" ? "." + Ot(d, 0) : H, Lt(w) ? (Y = "", $ != null && (Y = $.replace(Al, "$&/") + "/"), Bt(w, M, Y, "", function(El) {
        return El;
      })) : w != null && ($t(w) && (w = zt(
        w,
        Y + (w.key == null || d && d.key === w.key ? "" : ("" + w.key).replace(
          Al,
          "$&/"
        ) + "/") + $
      )), M.push(w)), 1;
    $ = 0;
    var el = H === "" ? "." : H + ":";
    if (Lt(d))
      for (var pt = 0; pt < d.length; pt++)
        H = d[pt], ut = el + Ot(H, pt), $ += Bt(
          H,
          M,
          Y,
          ut,
          w
        );
    else if (pt = W(d), typeof pt == "function")
      for (d = pt.call(d), pt = 0; !(H = d.next()).done; )
        H = H.value, ut = el + Ot(H, pt++), $ += Bt(
          H,
          M,
          Y,
          ut,
          w
        );
    else if (ut === "object") {
      if (typeof d.then == "function")
        return Bt(
          Il(d),
          M,
          Y,
          H,
          w
        );
      throw M = String(d), Error(
        "Objects are not valid as a React child (found: " + (M === "[object Object]" ? "object with keys {" + Object.keys(d).join(", ") + "}" : M) + "). If you meant to render a collection of children, use an array instead."
      );
    }
    return $;
  }
  function p(d, M, Y) {
    if (d == null) return d;
    var H = [], w = 0;
    return Bt(d, H, "", "", function(ut) {
      return M.call(Y, ut, w++);
    }), H;
  }
  function C(d) {
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
  var V = typeof reportError == "function" ? reportError : function(d) {
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
  function st() {
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
      if (!$t(d))
        throw Error(
          "React.Children.only expected to receive a single React element child."
        );
      return d;
    }
  }, I.Component = Dt, I.Fragment = z, I.Profiler = U, I.PureComponent = ft, I.StrictMode = r, I.Suspense = N, I.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = P, I.__COMPILER_RUNTIME = {
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
    var H = q({}, d.props), w = d.key, ut = void 0;
    if (M != null)
      for ($ in M.ref !== void 0 && (ut = void 0), M.key !== void 0 && (w = "" + M.key), M)
        !Vt.call(M, $) || $ === "key" || $ === "__self" || $ === "__source" || $ === "ref" && M.ref === void 0 || (H[$] = M[$]);
    var $ = arguments.length - 2;
    if ($ === 1) H.children = Y;
    else if (1 < $) {
      for (var el = Array($), pt = 0; pt < $; pt++)
        el[pt] = arguments[pt + 2];
      H.children = el;
    }
    return Z(d.type, w, void 0, void 0, ut, H);
  }, I.createContext = function(d) {
    return d = {
      $$typeof: _,
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
    var H, w = {}, ut = null;
    if (M != null)
      for (H in M.key !== void 0 && (ut = "" + M.key), M)
        Vt.call(M, H) && H !== "key" && H !== "__self" && H !== "__source" && (w[H] = M[H]);
    var $ = arguments.length - 2;
    if ($ === 1) w.children = Y;
    else if (1 < $) {
      for (var el = Array($), pt = 0; pt < $; pt++)
        el[pt] = arguments[pt + 2];
      w.children = el;
    }
    if (d && d.defaultProps)
      for (H in $ = d.defaultProps, $)
        w[H] === void 0 && (w[H] = $[H]);
    return Z(d, ut, void 0, void 0, null, w);
  }, I.createRef = function() {
    return { current: null };
  }, I.forwardRef = function(d) {
    return { $$typeof: O, render: d };
  }, I.isValidElement = $t, I.lazy = function(d) {
    return {
      $$typeof: B,
      _payload: { _status: -1, _result: d },
      _init: C
    };
  }, I.memo = function(d, M) {
    return {
      $$typeof: x,
      type: d,
      compare: M === void 0 ? null : M
    };
  }, I.startTransition = function(d) {
    var M = P.T, Y = {};
    P.T = Y;
    try {
      var H = d(), w = P.S;
      w !== null && w(Y, H), typeof H == "object" && H !== null && typeof H.then == "function" && H.then(st, V);
    } catch (ut) {
      V(ut);
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
    var H = P.H;
    if (typeof Y == "function")
      throw Error(
        "useEffect CRUD overload is not enabled in this build of React."
      );
    return H.useEffect(d, M);
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
var Md;
function _f() {
  return Md || (Md = 1, Df.exports = a1()), Df.exports;
}
var Of = { exports: {} }, ll = {};
var Nd;
function u1() {
  if (Nd) return ll;
  Nd = 1;
  var o = _f();
  function b(N) {
    var x = "https://react.dev/errors/" + N;
    if (1 < arguments.length) {
      x += "?args[]=" + encodeURIComponent(arguments[1]);
      for (var B = 2; B < arguments.length; B++)
        x += "&args[]=" + encodeURIComponent(arguments[B]);
    }
    return "Minified React error #" + N + "; visit " + x + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
  }
  function z() {
  }
  var r = {
    d: {
      f: z,
      r: function() {
        throw Error(b(522));
      },
      D: z,
      C: z,
      L: z,
      m: z,
      X: z,
      S: z,
      M: z
    },
    p: 0,
    findDOMNode: null
  }, U = /* @__PURE__ */ Symbol.for("react.portal");
  function D(N, x, B) {
    var X = 3 < arguments.length && arguments[3] !== void 0 ? arguments[3] : null;
    return {
      $$typeof: U,
      key: X == null ? null : "" + X,
      children: N,
      containerInfo: x,
      implementation: B
    };
  }
  var _ = o.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;
  function O(N, x) {
    if (N === "font") return "";
    if (typeof x == "string")
      return x === "use-credentials" ? x : "";
  }
  return ll.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE = r, ll.createPortal = function(N, x) {
    var B = 2 < arguments.length && arguments[2] !== void 0 ? arguments[2] : null;
    if (!x || x.nodeType !== 1 && x.nodeType !== 9 && x.nodeType !== 11)
      throw Error(b(299));
    return D(N, x, null, B);
  }, ll.flushSync = function(N) {
    var x = _.T, B = r.p;
    try {
      if (_.T = null, r.p = 2, N) return N();
    } finally {
      _.T = x, r.p = B, r.d.f();
    }
  }, ll.preconnect = function(N, x) {
    typeof N == "string" && (x ? (x = x.crossOrigin, x = typeof x == "string" ? x === "use-credentials" ? x : "" : void 0) : x = null, r.d.C(N, x));
  }, ll.prefetchDNS = function(N) {
    typeof N == "string" && r.d.D(N);
  }, ll.preinit = function(N, x) {
    if (typeof N == "string" && x && typeof x.as == "string") {
      var B = x.as, X = O(B, x.crossOrigin), W = typeof x.integrity == "string" ? x.integrity : void 0, it = typeof x.fetchPriority == "string" ? x.fetchPriority : void 0;
      B === "style" ? r.d.S(
        N,
        typeof x.precedence == "string" ? x.precedence : void 0,
        {
          crossOrigin: X,
          integrity: W,
          fetchPriority: it
        }
      ) : B === "script" && r.d.X(N, {
        crossOrigin: X,
        integrity: W,
        fetchPriority: it,
        nonce: typeof x.nonce == "string" ? x.nonce : void 0
      });
    }
  }, ll.preinitModule = function(N, x) {
    if (typeof N == "string")
      if (typeof x == "object" && x !== null) {
        if (x.as == null || x.as === "script") {
          var B = O(
            x.as,
            x.crossOrigin
          );
          r.d.M(N, {
            crossOrigin: B,
            integrity: typeof x.integrity == "string" ? x.integrity : void 0,
            nonce: typeof x.nonce == "string" ? x.nonce : void 0
          });
        }
      } else x == null && r.d.M(N);
  }, ll.preload = function(N, x) {
    if (typeof N == "string" && typeof x == "object" && x !== null && typeof x.as == "string") {
      var B = x.as, X = O(B, x.crossOrigin);
      r.d.L(N, B, {
        crossOrigin: X,
        integrity: typeof x.integrity == "string" ? x.integrity : void 0,
        nonce: typeof x.nonce == "string" ? x.nonce : void 0,
        type: typeof x.type == "string" ? x.type : void 0,
        fetchPriority: typeof x.fetchPriority == "string" ? x.fetchPriority : void 0,
        referrerPolicy: typeof x.referrerPolicy == "string" ? x.referrerPolicy : void 0,
        imageSrcSet: typeof x.imageSrcSet == "string" ? x.imageSrcSet : void 0,
        imageSizes: typeof x.imageSizes == "string" ? x.imageSizes : void 0,
        media: typeof x.media == "string" ? x.media : void 0
      });
    }
  }, ll.preloadModule = function(N, x) {
    if (typeof N == "string")
      if (x) {
        var B = O(x.as, x.crossOrigin);
        r.d.m(N, {
          as: typeof x.as == "string" && x.as !== "script" ? x.as : void 0,
          crossOrigin: B,
          integrity: typeof x.integrity == "string" ? x.integrity : void 0
        });
      } else r.d.m(N);
  }, ll.requestFormReset = function(N) {
    r.d.r(N);
  }, ll.unstable_batchedUpdates = function(N, x) {
    return N(x);
  }, ll.useFormState = function(N, x, B) {
    return _.H.useFormState(N, x, B);
  }, ll.useFormStatus = function() {
    return _.H.useHostTransitionStatus();
  }, ll.version = "19.1.1", ll;
}
var Dd;
function n1() {
  if (Dd) return Of.exports;
  Dd = 1;
  function o() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o);
      } catch (b) {
        console.error(b);
      }
  }
  return o(), Of.exports = u1(), Of.exports;
}
var Od;
function c1() {
  if (Od) return ku;
  Od = 1;
  var o = e1(), b = _f(), z = n1();
  function r(t) {
    var l = "https://react.dev/errors/" + t;
    if (1 < arguments.length) {
      l += "?args[]=" + encodeURIComponent(arguments[1]);
      for (var e = 2; e < arguments.length; e++)
        l += "&args[]=" + encodeURIComponent(arguments[e]);
    }
    return "Minified React error #" + t + "; visit " + l + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
  }
  function U(t) {
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
  function _(t) {
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
  function N(t) {
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
        for (var c = !1, f = u.child; f; ) {
          if (f === e) {
            c = !0, e = u, a = n;
            break;
          }
          if (f === a) {
            c = !0, a = u, e = n;
            break;
          }
          f = f.sibling;
        }
        if (!c) {
          for (f = n.child; f; ) {
            if (f === e) {
              c = !0, e = n, a = u;
              break;
            }
            if (f === a) {
              c = !0, a = n, e = u;
              break;
            }
            f = f.sibling;
          }
          if (!c) throw Error(r(189));
        }
      }
      if (e.alternate !== a) throw Error(r(190));
    }
    if (e.tag !== 3) throw Error(r(188));
    return e.stateNode.current === e ? t : l;
  }
  function x(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t;
    for (t = t.child; t !== null; ) {
      if (l = x(t), l !== null) return l;
      t = t.sibling;
    }
    return null;
  }
  var B = Object.assign, X = /* @__PURE__ */ Symbol.for("react.element"), W = /* @__PURE__ */ Symbol.for("react.transitional.element"), it = /* @__PURE__ */ Symbol.for("react.portal"), q = /* @__PURE__ */ Symbol.for("react.fragment"), ot = /* @__PURE__ */ Symbol.for("react.strict_mode"), Dt = /* @__PURE__ */ Symbol.for("react.profiler"), xt = /* @__PURE__ */ Symbol.for("react.provider"), ft = /* @__PURE__ */ Symbol.for("react.consumer"), Et = /* @__PURE__ */ Symbol.for("react.context"), Lt = /* @__PURE__ */ Symbol.for("react.forward_ref"), P = /* @__PURE__ */ Symbol.for("react.suspense"), Vt = /* @__PURE__ */ Symbol.for("react.suspense_list"), Z = /* @__PURE__ */ Symbol.for("react.memo"), zt = /* @__PURE__ */ Symbol.for("react.lazy"), $t = /* @__PURE__ */ Symbol.for("react.activity"), ge = /* @__PURE__ */ Symbol.for("react.memo_cache_sentinel"), Al = Symbol.iterator;
  function Ot(t) {
    return t === null || typeof t != "object" ? null : (t = Al && t[Al] || t["@@iterator"], typeof t == "function" ? t : null);
  }
  var vl = /* @__PURE__ */ Symbol.for("react.client.reference");
  function Il(t) {
    if (t == null) return null;
    if (typeof t == "function")
      return t.$$typeof === vl ? null : t.displayName || t.name || null;
    if (typeof t == "string") return t;
    switch (t) {
      case q:
        return "Fragment";
      case Dt:
        return "Profiler";
      case ot:
        return "StrictMode";
      case P:
        return "Suspense";
      case Vt:
        return "SuspenseList";
      case $t:
        return "Activity";
    }
    if (typeof t == "object")
      switch (t.$$typeof) {
        case it:
          return "Portal";
        case Et:
          return (t.displayName || "Context") + ".Provider";
        case ft:
          return (t._context.displayName || "Context") + ".Consumer";
        case Lt:
          var l = t.render;
          return t = t.displayName, t || (t = l.displayName || l.name || "", t = t !== "" ? "ForwardRef(" + t + ")" : "ForwardRef"), t;
        case Z:
          return l = t.displayName || null, l !== null ? l : Il(t.type) || "Memo";
        case zt:
          l = t._payload, t = t._init;
          try {
            return Il(t(l));
          } catch {
          }
      }
    return null;
  }
  var Bt = Array.isArray, p = b.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, C = z.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE, V = {
    pending: !1,
    data: null,
    method: null,
    action: null
  }, st = [], d = -1;
  function M(t) {
    return { current: t };
  }
  function Y(t) {
    0 > d || (t.current = st[d], st[d] = null, d--);
  }
  function H(t, l) {
    d++, st[d] = t.current, t.current = l;
  }
  var w = M(null), ut = M(null), $ = M(null), el = M(null);
  function pt(t, l) {
    switch (H($, l), H(ut, t), H(w, null), l.nodeType) {
      case 9:
      case 11:
        t = (t = l.documentElement) && (t = t.namespaceURI) ? Wo(t) : 0;
        break;
      default:
        if (t = l.tagName, l = l.namespaceURI)
          l = Wo(l), t = Fo(l, t);
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
    Y(w), H(w, t);
  }
  function El() {
    Y(w), Y(ut), Y($);
  }
  function Bl(t) {
    t.memoizedState !== null && H(el, t);
    var l = w.current, e = Fo(l, t.type);
    l !== e && (H(ut, t), H(w, e));
  }
  function Ze(t) {
    ut.current === t && (Y(w), Y(ut)), el.current === t && (Y(el), Zu._currentValue = V);
  }
  var te = Object.prototype.hasOwnProperty, fa = o.unstable_scheduleCallback, Fa = o.unstable_cancelCallback, Wu = o.unstable_shouldYield, ml = o.unstable_requestPaint, al = o.unstable_now, Se = o.unstable_getCurrentPriorityLevel, Zl = o.unstable_ImmediatePriority, cl = o.unstable_UserBlockingPriority, Ll = o.unstable_NormalPriority, Vl = o.unstable_LowPriority, Rt = o.unstable_IdlePriority, il = o.log, Pa = o.unstable_setDisableYieldValue, zl = null, nt = null;
  function Yl(t) {
    if (typeof il == "function" && Pa(t), nt && typeof nt.setStrictMode == "function")
      try {
        nt.setStrictMode(zl, t);
      } catch {
      }
  }
  var ul = Math.clz32 ? Math.clz32 : Pu, vc = Math.log, Fu = Math.LN2;
  function Pu(t) {
    return t >>>= 0, t === 0 ? 32 : 31 - (vc(t) / Fu | 0) | 0;
  }
  var sa = 256, ra = 4194304;
  function Ml(t) {
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
  function wl(t, l, e) {
    var a = t.pendingLanes;
    if (a === 0) return 0;
    var u = 0, n = t.suspendedLanes, c = t.pingedLanes;
    t = t.warmLanes;
    var f = a & 134217727;
    return f !== 0 ? (a = f & ~n, a !== 0 ? u = Ml(a) : (c &= f, c !== 0 ? u = Ml(c) : e || (e = f & ~t, e !== 0 && (u = Ml(e))))) : (f = a & ~n, f !== 0 ? u = Ml(f) : c !== 0 ? u = Ml(c) : e || (e = a & ~t, e !== 0 && (u = Ml(e)))), u === 0 ? 0 : l !== 0 && l !== u && (l & n) === 0 && (n = u & -u, e = l & -l, n >= e || n === 32 && (e & 4194048) !== 0) ? l : u;
  }
  function E(t, l) {
    return (t.pendingLanes & ~(t.suspendedLanes & ~t.pingedLanes) & l) === 0;
  }
  function Q(t, l) {
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
  function F() {
    var t = sa;
    return sa <<= 1, (sa & 4194048) === 0 && (sa = 256), t;
  }
  function jt() {
    var t = ra;
    return ra <<= 1, (ra & 62914560) === 0 && (ra = 4194304), t;
  }
  function At(t) {
    for (var l = [], e = 0; 31 > e; e++) l.push(t);
    return l;
  }
  function Ct(t, l) {
    t.pendingLanes |= l, l !== 268435456 && (t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0);
  }
  function oa(t, l, e, a, u, n) {
    var c = t.pendingLanes;
    t.pendingLanes = e, t.suspendedLanes = 0, t.pingedLanes = 0, t.warmLanes = 0, t.expiredLanes &= e, t.entangledLanes &= e, t.errorRecoveryDisabledLanes &= e, t.shellSuspendCounter = 0;
    var f = t.entanglements, s = t.expirationTimes, m = t.hiddenUpdates;
    for (e = c & ~e; 0 < e; ) {
      var j = 31 - ul(e), A = 1 << j;
      f[j] = 0, s[j] = -1;
      var g = m[j];
      if (g !== null)
        for (m[j] = null, j = 0; j < g.length; j++) {
          var S = g[j];
          S !== null && (S.lane &= -536870913);
        }
      e &= ~A;
    }
    a !== 0 && fl(t, a, 0), n !== 0 && u === 0 && t.tag !== 0 && (t.suspendedLanes |= n & ~(c & ~l));
  }
  function fl(t, l, e) {
    t.pendingLanes |= l, t.suspendedLanes &= ~l;
    var a = 31 - ul(l);
    t.entangledLanes |= l, t.entanglements[a] = t.entanglements[a] | 1073741824 | e & 4194090;
  }
  function Kl(t, l) {
    var e = t.entangledLanes |= l;
    for (t = t.entanglements; e; ) {
      var a = 31 - ul(e), u = 1 << a;
      u & l | t[a] & l && (t[a] |= l), e &= ~u;
    }
  }
  function Jl(t) {
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
  function Le(t) {
    return t &= -t, 2 < t ? 8 < t ? (t & 134217727) !== 0 ? 32 : 268435456 : 8 : 2;
  }
  function Iu() {
    var t = C.p;
    return t !== 0 ? t : (t = window.event, t === void 0 ? 32 : md(t.type));
  }
  function Zd(t, l) {
    var e = C.p;
    try {
      return C.p = t, l();
    } finally {
      C.p = e;
    }
  }
  var be = Math.random().toString(36).slice(2), It = "__reactFiber$" + be, sl = "__reactProps$" + be, da = "__reactContainer$" + be, mc = "__reactEvents$" + be, Ld = "__reactListeners$" + be, Vd = "__reactHandles$" + be, Bf = "__reactResources$" + be, Ia = "__reactMarker$" + be;
  function gc(t) {
    delete t[It], delete t[sl], delete t[mc], delete t[Ld], delete t[Vd];
  }
  function ha(t) {
    var l = t[It];
    if (l) return l;
    for (var e = t.parentNode; e; ) {
      if (l = e[da] || e[It]) {
        if (e = l.alternate, l.child !== null || e !== null && e.child !== null)
          for (t = ld(t); t !== null; ) {
            if (e = t[It]) return e;
            t = ld(t);
          }
        return l;
      }
      t = e, e = t.parentNode;
    }
    return null;
  }
  function ya(t) {
    if (t = t[It] || t[da]) {
      var l = t.tag;
      if (l === 5 || l === 6 || l === 13 || l === 26 || l === 27 || l === 3)
        return t;
    }
    return null;
  }
  function tu(t) {
    var l = t.tag;
    if (l === 5 || l === 26 || l === 27 || l === 6) return t.stateNode;
    throw Error(r(33));
  }
  function va(t) {
    var l = t[Bf];
    return l || (l = t[Bf] = { hoistableStyles: /* @__PURE__ */ new Map(), hoistableScripts: /* @__PURE__ */ new Map() }), l;
  }
  function wt(t) {
    t[Ia] = !0;
  }
  var Yf = /* @__PURE__ */ new Set(), Gf = {};
  function Ve(t, l) {
    ma(t, l), ma(t + "Capture", l);
  }
  function ma(t, l) {
    for (Gf[t] = l, t = 0; t < l.length; t++)
      Yf.add(l[t]);
  }
  var wd = RegExp(
    "^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"
  ), Xf = {}, Qf = {};
  function Kd(t) {
    return te.call(Qf, t) ? !0 : te.call(Xf, t) ? !1 : wd.test(t) ? Qf[t] = !0 : (Xf[t] = !0, !1);
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
  function le(t, l, e, a) {
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
  var Sc, Zf;
  function ga(t) {
    if (Sc === void 0)
      try {
        throw Error();
      } catch (e) {
        var l = e.stack.trim().match(/\n( *(at )?)/);
        Sc = l && l[1] || "", Zf = -1 < e.stack.indexOf(`
    at`) ? " (<anonymous>)" : -1 < e.stack.indexOf("@") ? "@unknown:0:0" : "";
      }
    return `
` + Sc + t + Zf;
  }
  var bc = !1;
  function xc(t, l) {
    if (!t || bc) return "";
    bc = !0;
    var e = Error.prepareStackTrace;
    Error.prepareStackTrace = void 0;
    try {
      var a = {
        DetermineComponentFrameRoot: function() {
          try {
            if (l) {
              var A = function() {
                throw Error();
              };
              if (Object.defineProperty(A.prototype, "props", {
                set: function() {
                  throw Error();
                }
              }), typeof Reflect == "object" && Reflect.construct) {
                try {
                  Reflect.construct(A, []);
                } catch (S) {
                  var g = S;
                }
                Reflect.construct(t, [], A);
              } else {
                try {
                  A.call();
                } catch (S) {
                  g = S;
                }
                t.call(A.prototype);
              }
            } else {
              try {
                throw Error();
              } catch (S) {
                g = S;
              }
              (A = t()) && typeof A.catch == "function" && A.catch(function() {
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
      var n = a.DetermineComponentFrameRoot(), c = n[0], f = n[1];
      if (c && f) {
        var s = c.split(`
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
      bc = !1, Error.prepareStackTrace = e;
    }
    return (e = t ? t.displayName || t.name : "") ? ga(e) : "";
  }
  function Jd(t) {
    switch (t.tag) {
      case 26:
      case 27:
      case 5:
        return ga(t.type);
      case 16:
        return ga("Lazy");
      case 13:
        return ga("Suspense");
      case 19:
        return ga("SuspenseList");
      case 0:
      case 15:
        return xc(t.type, !1);
      case 11:
        return xc(t.type.render, !1);
      case 1:
        return xc(t.type, !0);
      case 31:
        return ga("Activity");
      default:
        return "";
    }
  }
  function Lf(t) {
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
  function Vf(t) {
    var l = t.type;
    return (t = t.nodeName) && t.toLowerCase() === "input" && (l === "checkbox" || l === "radio");
  }
  function kd(t) {
    var l = Vf(t) ? "checked" : "value", e = Object.getOwnPropertyDescriptor(
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
        set: function(c) {
          a = "" + c, n.call(this, c);
        }
      }), Object.defineProperty(t, l, {
        enumerable: e.enumerable
      }), {
        getValue: function() {
          return a;
        },
        setValue: function(c) {
          a = "" + c;
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
  function wf(t) {
    if (!t) return !1;
    var l = t._valueTracker;
    if (!l) return !0;
    var e = l.getValue(), a = "";
    return t && (a = Vf(t) ? t.checked ? "true" : "false" : t.value), t = a, t !== e ? (l.setValue(t), !0) : !1;
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
  function pc(t, l, e, a, u, n, c, f) {
    t.name = "", c != null && typeof c != "function" && typeof c != "symbol" && typeof c != "boolean" ? t.type = c : t.removeAttribute("type"), l != null ? c === "number" ? (l === 0 && t.value === "" || t.value != l) && (t.value = "" + Nl(l)) : t.value !== "" + Nl(l) && (t.value = "" + Nl(l)) : c !== "submit" && c !== "reset" || t.removeAttribute("value"), l != null ? jc(t, c, Nl(l)) : e != null ? jc(t, c, Nl(e)) : a != null && t.removeAttribute("value"), u == null && n != null && (t.defaultChecked = !!n), u != null && (t.checked = u && typeof u != "function" && typeof u != "symbol"), f != null && typeof f != "function" && typeof f != "symbol" && typeof f != "boolean" ? t.name = "" + Nl(f) : t.removeAttribute("name");
  }
  function Kf(t, l, e, a, u, n, c, f) {
    if (n != null && typeof n != "function" && typeof n != "symbol" && typeof n != "boolean" && (t.type = n), l != null || e != null) {
      if (!(n !== "submit" && n !== "reset" || l != null))
        return;
      e = e != null ? "" + Nl(e) : "", l = l != null ? "" + Nl(l) : e, f || l === t.value || (t.value = l), t.defaultValue = l;
    }
    a = a ?? u, a = typeof a != "function" && typeof a != "symbol" && !!a, t.checked = f ? t.checked : !!a, t.defaultChecked = !!a, c != null && typeof c != "function" && typeof c != "symbol" && typeof c != "boolean" && (t.name = c);
  }
  function jc(t, l, e) {
    l === "number" && an(t.ownerDocument) === t || t.defaultValue === "" + e || (t.defaultValue = "" + e);
  }
  function Sa(t, l, e, a) {
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
  function Jf(t, l, e) {
    if (l != null && (l = "" + Nl(l), l !== t.value && (t.value = l), e == null)) {
      t.defaultValue !== l && (t.defaultValue = l);
      return;
    }
    t.defaultValue = e != null ? "" + Nl(e) : "";
  }
  function kf(t, l, e, a) {
    if (l == null) {
      if (a != null) {
        if (e != null) throw Error(r(92));
        if (Bt(a)) {
          if (1 < a.length) throw Error(r(93));
          a = a[0];
        }
        e = a;
      }
      e == null && (e = ""), l = e;
    }
    e = Nl(l), t.defaultValue = e, a = t.textContent, a === e && a !== "" && a !== null && (t.value = a);
  }
  function ba(t, l) {
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
  function $f(t, l, e) {
    var a = l.indexOf("--") === 0;
    e == null || typeof e == "boolean" || e === "" ? a ? t.setProperty(l, "") : l === "float" ? t.cssFloat = "" : t[l] = "" : a ? t.setProperty(l, e) : typeof e != "number" || e === 0 || Wd.has(l) ? l === "float" ? t.cssFloat = e : t[l] = ("" + e).trim() : t[l] = e + "px";
  }
  function Wf(t, l, e) {
    if (l != null && typeof l != "object")
      throw Error(r(62));
    if (t = t.style, e != null) {
      for (var a in e)
        !e.hasOwnProperty(a) || l != null && l.hasOwnProperty(a) || (a.indexOf("--") === 0 ? t.setProperty(a, "") : a === "float" ? t.cssFloat = "" : t[a] = "");
      for (var u in l)
        a = l[u], l.hasOwnProperty(u) && e[u] !== a && $f(t, u, a);
    } else
      for (var n in l)
        l.hasOwnProperty(n) && $f(t, n, l[n]);
  }
  function Tc(t) {
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
  var Ac = null;
  function Ec(t) {
    return t = t.target || t.srcElement || window, t.correspondingUseElement && (t = t.correspondingUseElement), t.nodeType === 3 ? t.parentNode : t;
  }
  var xa = null, pa = null;
  function Ff(t) {
    var l = ya(t);
    if (l && (t = l.stateNode)) {
      var e = t[sl] || null;
      t: switch (t = l.stateNode, l.type) {
        case "input":
          if (pc(
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
                if (!u) throw Error(r(90));
                pc(
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
              a = e[l], a.form === t.form && wf(a);
          }
          break t;
        case "textarea":
          Jf(t, e.value, e.defaultValue);
          break t;
        case "select":
          l = e.value, l != null && Sa(t, !!e.multiple, l, !1);
      }
    }
  }
  var zc = !1;
  function Pf(t, l, e) {
    if (zc) return t(l, e);
    zc = !0;
    try {
      var a = t(l);
      return a;
    } finally {
      if (zc = !1, (xa !== null || pa !== null) && (Vn(), xa && (l = xa, t = pa, pa = xa = null, Ff(l), t)))
        for (l = 0; l < t.length; l++) Ff(t[l]);
    }
  }
  function lu(t, l) {
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
        r(231, l, typeof e)
      );
    return e;
  }
  var ee = !(typeof window > "u" || typeof window.document > "u" || typeof window.document.createElement > "u"), Mc = !1;
  if (ee)
    try {
      var eu = {};
      Object.defineProperty(eu, "passive", {
        get: function() {
          Mc = !0;
        }
      }), window.addEventListener("test", eu, eu), window.removeEventListener("test", eu, eu);
    } catch {
      Mc = !1;
    }
  var xe = null, Nc = null, nn = null;
  function If() {
    if (nn) return nn;
    var t, l = Nc, e = l.length, a, u = "value" in xe ? xe.value : xe.textContent, n = u.length;
    for (t = 0; t < e && l[t] === u[t]; t++) ;
    var c = e - t;
    for (a = 1; a <= c && l[e - a] === u[n - a]; a++) ;
    return nn = u.slice(t, 1 < a ? 1 - a : void 0);
  }
  function cn(t) {
    var l = t.keyCode;
    return "charCode" in t ? (t = t.charCode, t === 0 && l === 13 && (t = 13)) : t = l, t === 10 && (t = 13), 32 <= t || t === 13 ? t : 0;
  }
  function fn() {
    return !0;
  }
  function ts() {
    return !1;
  }
  function rl(t) {
    function l(e, a, u, n, c) {
      this._reactName = e, this._targetInst = u, this.type = a, this.nativeEvent = n, this.target = c, this.currentTarget = null;
      for (var f in t)
        t.hasOwnProperty(f) && (e = t[f], this[f] = e ? e(n) : n[f]);
      return this.isDefaultPrevented = (n.defaultPrevented != null ? n.defaultPrevented : n.returnValue === !1) ? fn : ts, this.isPropagationStopped = ts, this;
    }
    return B(l.prototype, {
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
  var we = {
    eventPhase: 0,
    bubbles: 0,
    cancelable: 0,
    timeStamp: function(t) {
      return t.timeStamp || Date.now();
    },
    defaultPrevented: 0,
    isTrusted: 0
  }, sn = rl(we), au = B({}, we, { view: 0, detail: 0 }), Id = rl(au), Dc, Oc, uu, rn = B({}, au, {
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
    getModifierState: _c,
    button: 0,
    buttons: 0,
    relatedTarget: function(t) {
      return t.relatedTarget === void 0 ? t.fromElement === t.srcElement ? t.toElement : t.fromElement : t.relatedTarget;
    },
    movementX: function(t) {
      return "movementX" in t ? t.movementX : (t !== uu && (uu && t.type === "mousemove" ? (Dc = t.screenX - uu.screenX, Oc = t.screenY - uu.screenY) : Oc = Dc = 0, uu = t), Dc);
    },
    movementY: function(t) {
      return "movementY" in t ? t.movementY : Oc;
    }
  }), ls = rl(rn), th = B({}, rn, { dataTransfer: 0 }), lh = rl(th), eh = B({}, au, { relatedTarget: 0 }), Rc = rl(eh), ah = B({}, we, {
    animationName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), uh = rl(ah), nh = B({}, we, {
    clipboardData: function(t) {
      return "clipboardData" in t ? t.clipboardData : window.clipboardData;
    }
  }), ch = rl(nh), ih = B({}, we, { data: 0 }), es = rl(ih), fh = {
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
  function _c() {
    return oh;
  }
  var dh = B({}, au, {
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
    getModifierState: _c,
    charCode: function(t) {
      return t.type === "keypress" ? cn(t) : 0;
    },
    keyCode: function(t) {
      return t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    },
    which: function(t) {
      return t.type === "keypress" ? cn(t) : t.type === "keydown" || t.type === "keyup" ? t.keyCode : 0;
    }
  }), hh = rl(dh), yh = B({}, rn, {
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
  }), as = rl(yh), vh = B({}, au, {
    touches: 0,
    targetTouches: 0,
    changedTouches: 0,
    altKey: 0,
    metaKey: 0,
    ctrlKey: 0,
    shiftKey: 0,
    getModifierState: _c
  }), mh = rl(vh), gh = B({}, we, {
    propertyName: 0,
    elapsedTime: 0,
    pseudoElement: 0
  }), Sh = rl(gh), bh = B({}, rn, {
    deltaX: function(t) {
      return "deltaX" in t ? t.deltaX : "wheelDeltaX" in t ? -t.wheelDeltaX : 0;
    },
    deltaY: function(t) {
      return "deltaY" in t ? t.deltaY : "wheelDeltaY" in t ? -t.wheelDeltaY : "wheelDelta" in t ? -t.wheelDelta : 0;
    },
    deltaZ: 0,
    deltaMode: 0
  }), xh = rl(bh), ph = B({}, we, {
    newState: 0,
    oldState: 0
  }), jh = rl(ph), Th = [9, 13, 27, 32], Uc = ee && "CompositionEvent" in window, nu = null;
  ee && "documentMode" in document && (nu = document.documentMode);
  var Ah = ee && "TextEvent" in window && !nu, us = ee && (!Uc || nu && 8 < nu && 11 >= nu), ns = " ", cs = !1;
  function is(t, l) {
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
  function fs(t) {
    return t = t.detail, typeof t == "object" && "data" in t ? t.data : null;
  }
  var ja = !1;
  function Eh(t, l) {
    switch (t) {
      case "compositionend":
        return fs(l);
      case "keypress":
        return l.which !== 32 ? null : (cs = !0, ns);
      case "textInput":
        return t = l.data, t === ns && cs ? null : t;
      default:
        return null;
    }
  }
  function zh(t, l) {
    if (ja)
      return t === "compositionend" || !Uc && is(t, l) ? (t = If(), nn = Nc = xe = null, ja = !1, t) : null;
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
        return us && l.locale !== "ko" ? null : l.data;
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
  function ss(t) {
    var l = t && t.nodeName && t.nodeName.toLowerCase();
    return l === "input" ? !!Mh[t.type] : l === "textarea";
  }
  function rs(t, l, e, a) {
    xa ? pa ? pa.push(a) : pa = [a] : xa = a, l = Wn(l, "onChange"), 0 < l.length && (e = new sn(
      "onChange",
      "change",
      null,
      e,
      a
    ), t.push({ event: e, listeners: l }));
  }
  var cu = null, iu = null;
  function Nh(t) {
    wo(t, 0);
  }
  function on(t) {
    var l = tu(t);
    if (wf(l)) return t;
  }
  function os(t, l) {
    if (t === "change") return l;
  }
  var ds = !1;
  if (ee) {
    var Cc;
    if (ee) {
      var Hc = "oninput" in document;
      if (!Hc) {
        var hs = document.createElement("div");
        hs.setAttribute("oninput", "return;"), Hc = typeof hs.oninput == "function";
      }
      Cc = Hc;
    } else Cc = !1;
    ds = Cc && (!document.documentMode || 9 < document.documentMode);
  }
  function ys() {
    cu && (cu.detachEvent("onpropertychange", vs), iu = cu = null);
  }
  function vs(t) {
    if (t.propertyName === "value" && on(iu)) {
      var l = [];
      rs(
        l,
        iu,
        t,
        Ec(t)
      ), Pf(Nh, l);
    }
  }
  function Dh(t, l, e) {
    t === "focusin" ? (ys(), cu = l, iu = e, cu.attachEvent("onpropertychange", vs)) : t === "focusout" && ys();
  }
  function Oh(t) {
    if (t === "selectionchange" || t === "keyup" || t === "keydown")
      return on(iu);
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
  var gl = typeof Object.is == "function" ? Object.is : Uh;
  function fu(t, l) {
    if (gl(t, l)) return !0;
    if (typeof t != "object" || t === null || typeof l != "object" || l === null)
      return !1;
    var e = Object.keys(t), a = Object.keys(l);
    if (e.length !== a.length) return !1;
    for (a = 0; a < e.length; a++) {
      var u = e[a];
      if (!te.call(l, u) || !gl(t[u], l[u]))
        return !1;
    }
    return !0;
  }
  function ms(t) {
    for (; t && t.firstChild; ) t = t.firstChild;
    return t;
  }
  function gs(t, l) {
    var e = ms(t);
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
      e = ms(e);
    }
  }
  function Ss(t, l) {
    return t && l ? t === l ? !0 : t && t.nodeType === 3 ? !1 : l && l.nodeType === 3 ? Ss(t, l.parentNode) : "contains" in t ? t.contains(l) : t.compareDocumentPosition ? !!(t.compareDocumentPosition(l) & 16) : !1 : !1;
  }
  function bs(t) {
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
  function qc(t) {
    var l = t && t.nodeName && t.nodeName.toLowerCase();
    return l && (l === "input" && (t.type === "text" || t.type === "search" || t.type === "tel" || t.type === "url" || t.type === "password") || l === "textarea" || t.contentEditable === "true");
  }
  var Ch = ee && "documentMode" in document && 11 >= document.documentMode, Ta = null, Bc = null, su = null, Yc = !1;
  function xs(t, l, e) {
    var a = e.window === e ? e.document : e.nodeType === 9 ? e : e.ownerDocument;
    Yc || Ta == null || Ta !== an(a) || (a = Ta, "selectionStart" in a && qc(a) ? a = { start: a.selectionStart, end: a.selectionEnd } : (a = (a.ownerDocument && a.ownerDocument.defaultView || window).getSelection(), a = {
      anchorNode: a.anchorNode,
      anchorOffset: a.anchorOffset,
      focusNode: a.focusNode,
      focusOffset: a.focusOffset
    }), su && fu(su, a) || (su = a, a = Wn(Bc, "onSelect"), 0 < a.length && (l = new sn(
      "onSelect",
      "select",
      null,
      l,
      e
    ), t.push({ event: l, listeners: a }), l.target = Ta)));
  }
  function Ke(t, l) {
    var e = {};
    return e[t.toLowerCase()] = l.toLowerCase(), e["Webkit" + t] = "webkit" + l, e["Moz" + t] = "moz" + l, e;
  }
  var Aa = {
    animationend: Ke("Animation", "AnimationEnd"),
    animationiteration: Ke("Animation", "AnimationIteration"),
    animationstart: Ke("Animation", "AnimationStart"),
    transitionrun: Ke("Transition", "TransitionRun"),
    transitionstart: Ke("Transition", "TransitionStart"),
    transitioncancel: Ke("Transition", "TransitionCancel"),
    transitionend: Ke("Transition", "TransitionEnd")
  }, Gc = {}, ps = {};
  ee && (ps = document.createElement("div").style, "AnimationEvent" in window || (delete Aa.animationend.animation, delete Aa.animationiteration.animation, delete Aa.animationstart.animation), "TransitionEvent" in window || delete Aa.transitionend.transition);
  function Je(t) {
    if (Gc[t]) return Gc[t];
    if (!Aa[t]) return t;
    var l = Aa[t], e;
    for (e in l)
      if (l.hasOwnProperty(e) && e in ps)
        return Gc[t] = l[e];
    return t;
  }
  var js = Je("animationend"), Ts = Je("animationiteration"), As = Je("animationstart"), Hh = Je("transitionrun"), qh = Je("transitionstart"), Bh = Je("transitioncancel"), Es = Je("transitionend"), zs = /* @__PURE__ */ new Map(), Xc = "abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(
    " "
  );
  Xc.push("scrollEnd");
  function Gl(t, l) {
    zs.set(t, l), Ve(l, [t]);
  }
  var Ms = /* @__PURE__ */ new WeakMap();
  function Ol(t, l) {
    if (typeof t == "object" && t !== null) {
      var e = Ms.get(t);
      return e !== void 0 ? e : (l = {
        value: t,
        source: l,
        stack: Lf(l)
      }, Ms.set(t, l), l);
    }
    return {
      value: t,
      source: l,
      stack: Lf(l)
    };
  }
  var Rl = [], Ea = 0, Qc = 0;
  function dn() {
    for (var t = Ea, l = Qc = Ea = 0; l < t; ) {
      var e = Rl[l];
      Rl[l++] = null;
      var a = Rl[l];
      Rl[l++] = null;
      var u = Rl[l];
      Rl[l++] = null;
      var n = Rl[l];
      if (Rl[l++] = null, a !== null && u !== null) {
        var c = a.pending;
        c === null ? u.next = u : (u.next = c.next, c.next = u), a.pending = u;
      }
      n !== 0 && Ns(e, u, n);
    }
  }
  function hn(t, l, e, a) {
    Rl[Ea++] = t, Rl[Ea++] = l, Rl[Ea++] = e, Rl[Ea++] = a, Qc |= a, t.lanes |= a, t = t.alternate, t !== null && (t.lanes |= a);
  }
  function Zc(t, l, e, a) {
    return hn(t, l, e, a), yn(t);
  }
  function za(t, l) {
    return hn(t, null, null, l), yn(t);
  }
  function Ns(t, l, e) {
    t.lanes |= e;
    var a = t.alternate;
    a !== null && (a.lanes |= e);
    for (var u = !1, n = t.return; n !== null; )
      n.childLanes |= e, a = n.alternate, a !== null && (a.childLanes |= e), n.tag === 22 && (t = n.stateNode, t === null || t._visibility & 1 || (u = !0)), t = n, n = n.return;
    return t.tag === 3 ? (n = t.stateNode, u && l !== null && (u = 31 - ul(e), t = n.hiddenUpdates, a = t[u], a === null ? t[u] = [l] : a.push(l), l.lane = e | 536870912), n) : null;
  }
  function yn(t) {
    if (50 < Cu)
      throw Cu = 0, ki = null, Error(r(185));
    for (var l = t.return; l !== null; )
      t = l, l = t.return;
    return t.tag === 3 ? t.stateNode : null;
  }
  var Ma = {};
  function Yh(t, l, e, a) {
    this.tag = t, this.key = e, this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null, this.index = 0, this.refCleanup = this.ref = null, this.pendingProps = l, this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null, this.mode = a, this.subtreeFlags = this.flags = 0, this.deletions = null, this.childLanes = this.lanes = 0, this.alternate = null;
  }
  function Sl(t, l, e, a) {
    return new Yh(t, l, e, a);
  }
  function Lc(t) {
    return t = t.prototype, !(!t || !t.isReactComponent);
  }
  function ae(t, l) {
    var e = t.alternate;
    return e === null ? (e = Sl(
      t.tag,
      l,
      t.key,
      t.mode
    ), e.elementType = t.elementType, e.type = t.type, e.stateNode = t.stateNode, e.alternate = t, t.alternate = e) : (e.pendingProps = l, e.type = t.type, e.flags = 0, e.subtreeFlags = 0, e.deletions = null), e.flags = t.flags & 65011712, e.childLanes = t.childLanes, e.lanes = t.lanes, e.child = t.child, e.memoizedProps = t.memoizedProps, e.memoizedState = t.memoizedState, e.updateQueue = t.updateQueue, l = t.dependencies, e.dependencies = l === null ? null : { lanes: l.lanes, firstContext: l.firstContext }, e.sibling = t.sibling, e.index = t.index, e.ref = t.ref, e.refCleanup = t.refCleanup, e;
  }
  function Ds(t, l) {
    t.flags &= 65011714;
    var e = t.alternate;
    return e === null ? (t.childLanes = 0, t.lanes = l, t.child = null, t.subtreeFlags = 0, t.memoizedProps = null, t.memoizedState = null, t.updateQueue = null, t.dependencies = null, t.stateNode = null) : (t.childLanes = e.childLanes, t.lanes = e.lanes, t.child = e.child, t.subtreeFlags = 0, t.deletions = null, t.memoizedProps = e.memoizedProps, t.memoizedState = e.memoizedState, t.updateQueue = e.updateQueue, t.type = e.type, l = e.dependencies, t.dependencies = l === null ? null : {
      lanes: l.lanes,
      firstContext: l.firstContext
    }), t;
  }
  function vn(t, l, e, a, u, n) {
    var c = 0;
    if (a = t, typeof t == "function") Lc(t) && (c = 1);
    else if (typeof t == "string")
      c = X0(
        t,
        e,
        w.current
      ) ? 26 : t === "html" || t === "head" || t === "body" ? 27 : 5;
    else
      t: switch (t) {
        case $t:
          return t = Sl(31, e, l, u), t.elementType = $t, t.lanes = n, t;
        case q:
          return ke(e.children, u, n, l);
        case ot:
          c = 8, u |= 24;
          break;
        case Dt:
          return t = Sl(12, e, l, u | 2), t.elementType = Dt, t.lanes = n, t;
        case P:
          return t = Sl(13, e, l, u), t.elementType = P, t.lanes = n, t;
        case Vt:
          return t = Sl(19, e, l, u), t.elementType = Vt, t.lanes = n, t;
        default:
          if (typeof t == "object" && t !== null)
            switch (t.$$typeof) {
              case xt:
              case Et:
                c = 10;
                break t;
              case ft:
                c = 9;
                break t;
              case Lt:
                c = 11;
                break t;
              case Z:
                c = 14;
                break t;
              case zt:
                c = 16, a = null;
                break t;
            }
          c = 29, e = Error(
            r(130, t === null ? "null" : typeof t, "")
          ), a = null;
      }
    return l = Sl(c, e, l, u), l.elementType = t, l.type = a, l.lanes = n, l;
  }
  function ke(t, l, e, a) {
    return t = Sl(7, t, a, l), t.lanes = e, t;
  }
  function Vc(t, l, e) {
    return t = Sl(6, t, null, l), t.lanes = e, t;
  }
  function wc(t, l, e) {
    return l = Sl(
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
  var Na = [], Da = 0, mn = null, gn = 0, _l = [], Ul = 0, $e = null, ue = 1, ne = "";
  function We(t, l) {
    Na[Da++] = gn, Na[Da++] = mn, mn = t, gn = l;
  }
  function Os(t, l, e) {
    _l[Ul++] = ue, _l[Ul++] = ne, _l[Ul++] = $e, $e = t;
    var a = ue;
    t = ne;
    var u = 32 - ul(a) - 1;
    a &= ~(1 << u), e += 1;
    var n = 32 - ul(l) + u;
    if (30 < n) {
      var c = u - u % 5;
      n = (a & (1 << c) - 1).toString(32), a >>= c, u -= c, ue = 1 << 32 - ul(l) + u | e << u | a, ne = n + t;
    } else
      ue = 1 << n | e << u | a, ne = t;
  }
  function Kc(t) {
    t.return !== null && (We(t, 1), Os(t, 1, 0));
  }
  function Jc(t) {
    for (; t === mn; )
      mn = Na[--Da], Na[Da] = null, gn = Na[--Da], Na[Da] = null;
    for (; t === $e; )
      $e = _l[--Ul], _l[Ul] = null, ne = _l[--Ul], _l[Ul] = null, ue = _l[--Ul], _l[Ul] = null;
  }
  var nl = null, _t = null, dt = !1, Fe = null, kl = !1, kc = Error(r(519));
  function Pe(t) {
    var l = Error(r(418, ""));
    throw du(Ol(l, t)), kc;
  }
  function Rs(t) {
    var l = t.stateNode, e = t.type, a = t.memoizedProps;
    switch (l[It] = t, l[sl] = a, e) {
      case "dialog":
        at("cancel", l), at("close", l);
        break;
      case "iframe":
      case "object":
      case "embed":
        at("load", l);
        break;
      case "video":
      case "audio":
        for (e = 0; e < qu.length; e++)
          at(qu[e], l);
        break;
      case "source":
        at("error", l);
        break;
      case "img":
      case "image":
      case "link":
        at("error", l), at("load", l);
        break;
      case "details":
        at("toggle", l);
        break;
      case "input":
        at("invalid", l), Kf(
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
        at("invalid", l);
        break;
      case "textarea":
        at("invalid", l), kf(l, a.value, a.defaultValue, a.children), en(l);
    }
    e = a.children, typeof e != "string" && typeof e != "number" && typeof e != "bigint" || l.textContent === "" + e || a.suppressHydrationWarning === !0 || $o(l.textContent, e) ? (a.popover != null && (at("beforetoggle", l), at("toggle", l)), a.onScroll != null && at("scroll", l), a.onScrollEnd != null && at("scrollend", l), a.onClick != null && (l.onclick = Fn), l = !0) : l = !1, l || Pe(t);
  }
  function _s(t) {
    for (nl = t.return; nl; )
      switch (nl.tag) {
        case 5:
        case 13:
          kl = !1;
          return;
        case 27:
        case 3:
          kl = !0;
          return;
        default:
          nl = nl.return;
      }
  }
  function ru(t) {
    if (t !== nl) return !1;
    if (!dt) return _s(t), dt = !0, !1;
    var l = t.tag, e;
    if ((e = l !== 3 && l !== 27) && ((e = l === 5) && (e = t.type, e = !(e !== "form" && e !== "button") || of(t.type, t.memoizedProps)), e = !e), e && _t && Pe(t), _s(t), l === 13) {
      if (t = t.memoizedState, t = t !== null ? t.dehydrated : null, !t) throw Error(r(317));
      t: {
        for (t = t.nextSibling, l = 0; t; ) {
          if (t.nodeType === 8)
            if (e = t.data, e === "/$") {
              if (l === 0) {
                _t = Ql(t.nextSibling);
                break t;
              }
              l--;
            } else
              e !== "$" && e !== "$!" && e !== "$?" || l++;
          t = t.nextSibling;
        }
        _t = null;
      }
    } else
      l === 27 ? (l = _t, qe(t.type) ? (t = vf, vf = null, _t = t) : _t = l) : _t = nl ? Ql(t.stateNode.nextSibling) : null;
    return !0;
  }
  function ou() {
    _t = nl = null, dt = !1;
  }
  function Us() {
    var t = Fe;
    return t !== null && (hl === null ? hl = t : hl.push.apply(
      hl,
      t
    ), Fe = null), t;
  }
  function du(t) {
    Fe === null ? Fe = [t] : Fe.push(t);
  }
  var $c = M(null), Ie = null, ce = null;
  function pe(t, l, e) {
    H($c, l._currentValue), l._currentValue = e;
  }
  function ie(t) {
    t._currentValue = $c.current, Y($c);
  }
  function Wc(t, l, e) {
    for (; t !== null; ) {
      var a = t.alternate;
      if ((t.childLanes & l) !== l ? (t.childLanes |= l, a !== null && (a.childLanes |= l)) : a !== null && (a.childLanes & l) !== l && (a.childLanes |= l), t === e) break;
      t = t.return;
    }
  }
  function Fc(t, l, e, a) {
    var u = t.child;
    for (u !== null && (u.return = t); u !== null; ) {
      var n = u.dependencies;
      if (n !== null) {
        var c = u.child;
        n = n.firstContext;
        t: for (; n !== null; ) {
          var f = n;
          n = u;
          for (var s = 0; s < l.length; s++)
            if (f.context === l[s]) {
              n.lanes |= e, f = n.alternate, f !== null && (f.lanes |= e), Wc(
                n.return,
                e,
                t
              ), a || (c = null);
              break t;
            }
          n = f.next;
        }
      } else if (u.tag === 18) {
        if (c = u.return, c === null) throw Error(r(341));
        c.lanes |= e, n = c.alternate, n !== null && (n.lanes |= e), Wc(c, e, t), c = null;
      } else c = u.child;
      if (c !== null) c.return = u;
      else
        for (c = u; c !== null; ) {
          if (c === t) {
            c = null;
            break;
          }
          if (u = c.sibling, u !== null) {
            u.return = c.return, c = u;
            break;
          }
          c = c.return;
        }
      u = c;
    }
  }
  function hu(t, l, e, a) {
    t = null;
    for (var u = l, n = !1; u !== null; ) {
      if (!n) {
        if ((u.flags & 524288) !== 0) n = !0;
        else if ((u.flags & 262144) !== 0) break;
      }
      if (u.tag === 10) {
        var c = u.alternate;
        if (c === null) throw Error(r(387));
        if (c = c.memoizedProps, c !== null) {
          var f = u.type;
          gl(u.pendingProps.value, c.value) || (t !== null ? t.push(f) : t = [f]);
        }
      } else if (u === el.current) {
        if (c = u.alternate, c === null) throw Error(r(387));
        c.memoizedState.memoizedState !== u.memoizedState.memoizedState && (t !== null ? t.push(Zu) : t = [Zu]);
      }
      u = u.return;
    }
    t !== null && Fc(
      l,
      t,
      e,
      a
    ), l.flags |= 262144;
  }
  function Sn(t) {
    for (t = t.firstContext; t !== null; ) {
      if (!gl(
        t.context._currentValue,
        t.memoizedValue
      ))
        return !0;
      t = t.next;
    }
    return !1;
  }
  function ta(t) {
    Ie = t, ce = null, t = t.dependencies, t !== null && (t.firstContext = null);
  }
  function tl(t) {
    return Cs(Ie, t);
  }
  function bn(t, l) {
    return Ie === null && ta(t), Cs(t, l);
  }
  function Cs(t, l) {
    var e = l._currentValue;
    if (l = { context: l, memoizedValue: e, next: null }, ce === null) {
      if (t === null) throw Error(r(308));
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
  }, Xh = o.unstable_scheduleCallback, Qh = o.unstable_NormalPriority, Xt = {
    $$typeof: Et,
    Consumer: null,
    Provider: null,
    _currentValue: null,
    _currentValue2: null,
    _threadCount: 0
  };
  function Pc() {
    return {
      controller: new Gh(),
      data: /* @__PURE__ */ new Map(),
      refCount: 0
    };
  }
  function yu(t) {
    t.refCount--, t.refCount === 0 && Xh(Qh, function() {
      t.controller.abort();
    });
  }
  var vu = null, Ic = 0, Oa = 0, Ra = null;
  function Zh(t, l) {
    if (vu === null) {
      var e = vu = [];
      Ic = 0, Oa = lf(), Ra = {
        status: "pending",
        value: void 0,
        then: function(a) {
          e.push(a);
        }
      };
    }
    return Ic++, l.then(Hs, Hs), l;
  }
  function Hs() {
    if (--Ic === 0 && vu !== null) {
      Ra !== null && (Ra.status = "fulfilled");
      var t = vu;
      vu = null, Oa = 0, Ra = null;
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
  var qs = p.S;
  p.S = function(t, l) {
    typeof l == "object" && l !== null && typeof l.then == "function" && Zh(t, l), qs !== null && qs(t, l);
  };
  var la = M(null);
  function ti() {
    var t = la.current;
    return t !== null ? t : Tt.pooledCache;
  }
  function xn(t, l) {
    l === null ? H(la, la.current) : H(la, l.pool);
  }
  function Bs() {
    var t = ti();
    return t === null ? null : { parent: Xt._currentValue, pool: t };
  }
  var mu = Error(r(460)), Ys = Error(r(474)), pn = Error(r(542)), li = { then: function() {
  } };
  function Gs(t) {
    return t = t.status, t === "fulfilled" || t === "rejected";
  }
  function jn() {
  }
  function Xs(t, l, e) {
    switch (e = t[e], e === void 0 ? t.push(l) : e !== l && (l.then(jn, jn), l = e), l.status) {
      case "fulfilled":
        return l.value;
      case "rejected":
        throw t = l.reason, Zs(t), t;
      default:
        if (typeof l.status == "string") l.then(jn, jn);
        else {
          if (t = Tt, t !== null && 100 < t.shellSuspendCounter)
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
            throw t = l.reason, Zs(t), t;
        }
        throw gu = l, mu;
    }
  }
  var gu = null;
  function Qs() {
    if (gu === null) throw Error(r(459));
    var t = gu;
    return gu = null, t;
  }
  function Zs(t) {
    if (t === mu || t === pn)
      throw Error(r(483));
  }
  var je = !1;
  function ei(t) {
    t.updateQueue = {
      baseState: t.memoizedState,
      firstBaseUpdate: null,
      lastBaseUpdate: null,
      shared: { pending: null, lanes: 0, hiddenCallbacks: null },
      callbacks: null
    };
  }
  function ai(t, l) {
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
    if (a = a.shared, (ht & 2) !== 0) {
      var u = a.pending;
      return u === null ? l.next = l : (l.next = u.next, u.next = l), a.pending = l, l = yn(t), Ns(t, null, e), l;
    }
    return hn(t, a, l, e), yn(t);
  }
  function Su(t, l, e) {
    if (l = l.updateQueue, l !== null && (l = l.shared, (e & 4194048) !== 0)) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, Kl(t, e);
    }
  }
  function ui(t, l) {
    var e = t.updateQueue, a = t.alternate;
    if (a !== null && (a = a.updateQueue, e === a)) {
      var u = null, n = null;
      if (e = e.firstBaseUpdate, e !== null) {
        do {
          var c = {
            lane: e.lane,
            tag: e.tag,
            payload: e.payload,
            callback: null,
            next: null
          };
          n === null ? u = n = c : n = n.next = c, e = e.next;
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
  var ni = !1;
  function bu() {
    if (ni) {
      var t = Ra;
      if (t !== null) throw t;
    }
  }
  function xu(t, l, e, a) {
    ni = !1;
    var u = t.updateQueue;
    je = !1;
    var n = u.firstBaseUpdate, c = u.lastBaseUpdate, f = u.shared.pending;
    if (f !== null) {
      u.shared.pending = null;
      var s = f, m = s.next;
      s.next = null, c === null ? n = m : c.next = m, c = s;
      var j = t.alternate;
      j !== null && (j = j.updateQueue, f = j.lastBaseUpdate, f !== c && (f === null ? j.firstBaseUpdate = m : f.next = m, j.lastBaseUpdate = s));
    }
    if (n !== null) {
      var A = u.baseState;
      c = 0, j = m = s = null, f = n;
      do {
        var g = f.lane & -536870913, S = g !== f.lane;
        if (S ? (ct & g) === g : (a & g) === g) {
          g !== 0 && g === Oa && (ni = !0), j !== null && (j = j.next = {
            lane: 0,
            tag: f.tag,
            payload: f.payload,
            callback: null,
            next: null
          });
          t: {
            var k = t, K = f;
            g = l;
            var gt = e;
            switch (K.tag) {
              case 1:
                if (k = K.payload, typeof k == "function") {
                  A = k.call(gt, A, g);
                  break t;
                }
                A = k;
                break t;
              case 3:
                k.flags = k.flags & -65537 | 128;
              case 0:
                if (k = K.payload, g = typeof k == "function" ? k.call(gt, A, g) : k, g == null) break t;
                A = B({}, A, g);
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
          }, j === null ? (m = j = S, s = A) : j = j.next = S, c |= g;
        if (f = f.next, f === null) {
          if (f = u.shared.pending, f === null)
            break;
          S = f, f = S.next, S.next = null, u.lastBaseUpdate = S, u.shared.pending = null;
        }
      } while (!0);
      j === null && (s = A), u.baseState = s, u.firstBaseUpdate = m, u.lastBaseUpdate = j, n === null && (u.shared.lanes = 0), _e |= c, t.lanes = c, t.memoizedState = A;
    }
  }
  function Ls(t, l) {
    if (typeof t != "function")
      throw Error(r(191, t));
    t.call(l);
  }
  function Vs(t, l) {
    var e = t.callbacks;
    if (e !== null)
      for (t.callbacks = null, t = 0; t < e.length; t++)
        Ls(e[t], l);
  }
  var _a = M(null), Tn = M(0);
  function ws(t, l) {
    t = ye, H(Tn, t), H(_a, l), ye = t | l.baseLanes;
  }
  function ci() {
    H(Tn, ye), H(_a, _a.current);
  }
  function ii() {
    ye = Tn.current, Y(_a), Y(Tn);
  }
  var Ee = 0, tt = null, vt = null, Yt = null, An = !1, Ua = !1, ea = !1, En = 0, pu = 0, Ca = null, Vh = 0;
  function Ht() {
    throw Error(r(321));
  }
  function fi(t, l) {
    if (l === null) return !1;
    for (var e = 0; e < l.length && e < t.length; e++)
      if (!gl(t[e], l[e])) return !1;
    return !0;
  }
  function si(t, l, e, a, u, n) {
    return Ee = n, tt = l, l.memoizedState = null, l.updateQueue = null, l.lanes = 0, p.H = t === null || t.memoizedState === null ? Nr : Dr, ea = !1, n = e(a, u), ea = !1, Ua && (n = Js(
      l,
      e,
      a,
      u
    )), Ks(t), n;
  }
  function Ks(t) {
    p.H = Rn;
    var l = vt !== null && vt.next !== null;
    if (Ee = 0, Yt = vt = tt = null, An = !1, pu = 0, Ca = null, l) throw Error(r(300));
    t === null || Kt || (t = t.dependencies, t !== null && Sn(t) && (Kt = !0));
  }
  function Js(t, l, e, a) {
    tt = t;
    var u = 0;
    do {
      if (Ua && (Ca = null), pu = 0, Ua = !1, 25 <= u) throw Error(r(301));
      if (u += 1, Yt = vt = null, t.updateQueue != null) {
        var n = t.updateQueue;
        n.lastEffect = null, n.events = null, n.stores = null, n.memoCache != null && (n.memoCache.index = 0);
      }
      p.H = Fh, n = l(e, a);
    } while (Ua);
    return n;
  }
  function wh() {
    var t = p.H, l = t.useState()[0];
    return l = typeof l.then == "function" ? ju(l) : l, t = t.useState()[0], (vt !== null ? vt.memoizedState : null) !== t && (tt.flags |= 1024), l;
  }
  function ri() {
    var t = En !== 0;
    return En = 0, t;
  }
  function oi(t, l, e) {
    l.updateQueue = t.updateQueue, l.flags &= -2053, t.lanes &= ~e;
  }
  function di(t) {
    if (An) {
      for (t = t.memoizedState; t !== null; ) {
        var l = t.queue;
        l !== null && (l.pending = null), t = t.next;
      }
      An = !1;
    }
    Ee = 0, Yt = vt = tt = null, Ua = !1, pu = En = 0, Ca = null;
  }
  function ol() {
    var t = {
      memoizedState: null,
      baseState: null,
      baseQueue: null,
      queue: null,
      next: null
    };
    return Yt === null ? tt.memoizedState = Yt = t : Yt = Yt.next = t, Yt;
  }
  function Gt() {
    if (vt === null) {
      var t = tt.alternate;
      t = t !== null ? t.memoizedState : null;
    } else t = vt.next;
    var l = Yt === null ? tt.memoizedState : Yt.next;
    if (l !== null)
      Yt = l, vt = t;
    else {
      if (t === null)
        throw tt.alternate === null ? Error(r(467)) : Error(r(310));
      vt = t, t = {
        memoizedState: vt.memoizedState,
        baseState: vt.baseState,
        baseQueue: vt.baseQueue,
        queue: vt.queue,
        next: null
      }, Yt === null ? tt.memoizedState = Yt = t : Yt = Yt.next = t;
    }
    return Yt;
  }
  function hi() {
    return { lastEffect: null, events: null, stores: null, memoCache: null };
  }
  function ju(t) {
    var l = pu;
    return pu += 1, Ca === null && (Ca = []), t = Xs(Ca, t, l), l = tt, (Yt === null ? l.memoizedState : Yt.next) === null && (l = l.alternate, p.H = l === null || l.memoizedState === null ? Nr : Dr), t;
  }
  function zn(t) {
    if (t !== null && typeof t == "object") {
      if (typeof t.then == "function") return ju(t);
      if (t.$$typeof === Et) return tl(t);
    }
    throw Error(r(438, String(t)));
  }
  function yi(t) {
    var l = null, e = tt.updateQueue;
    if (e !== null && (l = e.memoCache), l == null) {
      var a = tt.alternate;
      a !== null && (a = a.updateQueue, a !== null && (a = a.memoCache, a != null && (l = {
        data: a.data.map(function(u) {
          return u.slice();
        }),
        index: 0
      })));
    }
    if (l == null && (l = { data: [], index: 0 }), e === null && (e = hi(), tt.updateQueue = e), e.memoCache = l, e = l.data[l.index], e === void 0)
      for (e = l.data[l.index] = Array(t), a = 0; a < t; a++)
        e[a] = ge;
    return l.index++, e;
  }
  function fe(t, l) {
    return typeof l == "function" ? l(t) : l;
  }
  function Mn(t) {
    var l = Gt();
    return vi(l, vt, t);
  }
  function vi(t, l, e) {
    var a = t.queue;
    if (a === null) throw Error(r(311));
    a.lastRenderedReducer = e;
    var u = t.baseQueue, n = a.pending;
    if (n !== null) {
      if (u !== null) {
        var c = u.next;
        u.next = n.next, n.next = c;
      }
      l.baseQueue = u = n, a.pending = null;
    }
    if (n = t.baseState, u === null) t.memoizedState = n;
    else {
      l = u.next;
      var f = c = null, s = null, m = l, j = !1;
      do {
        var A = m.lane & -536870913;
        if (A !== m.lane ? (ct & A) === A : (Ee & A) === A) {
          var g = m.revertLane;
          if (g === 0)
            s !== null && (s = s.next = {
              lane: 0,
              revertLane: 0,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }), A === Oa && (j = !0);
          else if ((Ee & g) === g) {
            m = m.next, g === Oa && (j = !0);
            continue;
          } else
            A = {
              lane: 0,
              revertLane: m.revertLane,
              action: m.action,
              hasEagerState: m.hasEagerState,
              eagerState: m.eagerState,
              next: null
            }, s === null ? (f = s = A, c = n) : s = s.next = A, tt.lanes |= g, _e |= g;
          A = m.action, ea && e(n, A), n = m.hasEagerState ? m.eagerState : e(n, A);
        } else
          g = {
            lane: A,
            revertLane: m.revertLane,
            action: m.action,
            hasEagerState: m.hasEagerState,
            eagerState: m.eagerState,
            next: null
          }, s === null ? (f = s = g, c = n) : s = s.next = g, tt.lanes |= A, _e |= A;
        m = m.next;
      } while (m !== null && m !== l);
      if (s === null ? c = n : s.next = f, !gl(n, t.memoizedState) && (Kt = !0, j && (e = Ra, e !== null)))
        throw e;
      t.memoizedState = n, t.baseState = c, t.baseQueue = s, a.lastRenderedState = n;
    }
    return u === null && (a.lanes = 0), [t.memoizedState, a.dispatch];
  }
  function mi(t) {
    var l = Gt(), e = l.queue;
    if (e === null) throw Error(r(311));
    e.lastRenderedReducer = t;
    var a = e.dispatch, u = e.pending, n = l.memoizedState;
    if (u !== null) {
      e.pending = null;
      var c = u = u.next;
      do
        n = t(n, c.action), c = c.next;
      while (c !== u);
      gl(n, l.memoizedState) || (Kt = !0), l.memoizedState = n, l.baseQueue === null && (l.baseState = n), e.lastRenderedState = n;
    }
    return [n, a];
  }
  function ks(t, l, e) {
    var a = tt, u = Gt(), n = dt;
    if (n) {
      if (e === void 0) throw Error(r(407));
      e = e();
    } else e = l();
    var c = !gl(
      (vt || u).memoizedState,
      e
    );
    c && (u.memoizedState = e, Kt = !0), u = u.queue;
    var f = Fs.bind(null, a, u, t);
    if (Tu(2048, 8, f, [t]), u.getSnapshot !== l || c || Yt !== null && Yt.memoizedState.tag & 1) {
      if (a.flags |= 2048, Ha(
        9,
        Nn(),
        Ws.bind(
          null,
          a,
          u,
          e,
          l
        ),
        null
      ), Tt === null) throw Error(r(349));
      n || (Ee & 124) !== 0 || $s(a, l, e);
    }
    return e;
  }
  function $s(t, l, e) {
    t.flags |= 16384, t = { getSnapshot: l, value: e }, l = tt.updateQueue, l === null ? (l = hi(), tt.updateQueue = l, l.stores = [t]) : (e = l.stores, e === null ? l.stores = [t] : e.push(t));
  }
  function Ws(t, l, e, a) {
    l.value = e, l.getSnapshot = a, Ps(l) && Is(t);
  }
  function Fs(t, l, e) {
    return e(function() {
      Ps(l) && Is(t);
    });
  }
  function Ps(t) {
    var l = t.getSnapshot;
    t = t.value;
    try {
      var e = l();
      return !gl(t, e);
    } catch {
      return !0;
    }
  }
  function Is(t) {
    var l = za(t, 2);
    l !== null && Tl(l, t, 2);
  }
  function gi(t) {
    var l = ol();
    if (typeof t == "function") {
      var e = t;
      if (t = e(), ea) {
        Yl(!0);
        try {
          e();
        } finally {
          Yl(!1);
        }
      }
    }
    return l.memoizedState = l.baseState = t, l.queue = {
      pending: null,
      lanes: 0,
      dispatch: null,
      lastRenderedReducer: fe,
      lastRenderedState: t
    }, l;
  }
  function tr(t, l, e, a) {
    return t.baseState = e, vi(
      t,
      vt,
      typeof a == "function" ? a : fe
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
        then: function(c) {
          n.listeners.push(c);
        }
      };
      p.T !== null ? e(!0) : n.isTransition = !1, a(n), e = l.pending, e === null ? (n.next = l.pending = n, lr(l, n)) : (n.next = e.next, l.pending = e.next = n);
    }
  }
  function lr(t, l) {
    var e = l.action, a = l.payload, u = t.state;
    if (l.isTransition) {
      var n = p.T, c = {};
      p.T = c;
      try {
        var f = e(u, a), s = p.S;
        s !== null && s(c, f), er(t, l, f);
      } catch (m) {
        Si(t, l, m);
      } finally {
        p.T = n;
      }
    } else
      try {
        n = e(u, a), er(t, l, n);
      } catch (m) {
        Si(t, l, m);
      }
  }
  function er(t, l, e) {
    e !== null && typeof e == "object" && typeof e.then == "function" ? e.then(
      function(a) {
        ar(t, l, a);
      },
      function(a) {
        return Si(t, l, a);
      }
    ) : ar(t, l, e);
  }
  function ar(t, l, e) {
    l.status = "fulfilled", l.value = e, ur(l), t.state = e, l = t.pending, l !== null && (e = l.next, e === l ? t.pending = null : (e = e.next, l.next = e, lr(t, e)));
  }
  function Si(t, l, e) {
    var a = t.pending;
    if (t.pending = null, a !== null) {
      a = a.next;
      do
        l.status = "rejected", l.reason = e, ur(l), l = l.next;
      while (l !== a);
    }
    t.action = null;
  }
  function ur(t) {
    t = t.listeners;
    for (var l = 0; l < t.length; l++) (0, t[l])();
  }
  function nr(t, l) {
    return l;
  }
  function cr(t, l) {
    if (dt) {
      var e = Tt.formState;
      if (e !== null) {
        t: {
          var a = tt;
          if (dt) {
            if (_t) {
              l: {
                for (var u = _t, n = kl; u.nodeType !== 8; ) {
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
                _t = Ql(
                  u.nextSibling
                ), a = u.data === "F!";
                break t;
              }
            }
            Pe(a);
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
      lastRenderedReducer: nr,
      lastRenderedState: l
    }, e.queue = a, e = Er.bind(
      null,
      tt,
      a
    ), a.dispatch = e, a = gi(!1), n = Ti.bind(
      null,
      tt,
      !1,
      a.queue
    ), a = ol(), u = {
      state: l,
      dispatch: null,
      action: t,
      pending: null
    }, a.queue = u, e = Kh.bind(
      null,
      tt,
      u,
      n,
      e
    ), u.dispatch = e, a.memoizedState = t, [l, e, !1];
  }
  function ir(t) {
    var l = Gt();
    return fr(l, vt, t);
  }
  function fr(t, l, e) {
    if (l = vi(
      t,
      l,
      nr
    )[0], t = Mn(fe)[0], typeof l == "object" && l !== null && typeof l.then == "function")
      try {
        var a = ju(l);
      } catch (c) {
        throw c === mu ? pn : c;
      }
    else a = l;
    l = Gt();
    var u = l.queue, n = u.dispatch;
    return e !== l.memoizedState && (tt.flags |= 2048, Ha(
      9,
      Nn(),
      Jh.bind(null, u, e),
      null
    )), [a, n, t];
  }
  function Jh(t, l) {
    t.action = l;
  }
  function sr(t) {
    var l = Gt(), e = vt;
    if (e !== null)
      return fr(l, e, t);
    Gt(), l = l.memoizedState, e = Gt();
    var a = e.queue.dispatch;
    return e.memoizedState = t, [l, a, !1];
  }
  function Ha(t, l, e, a) {
    return t = { tag: t, create: e, deps: a, inst: l, next: null }, l = tt.updateQueue, l === null && (l = hi(), tt.updateQueue = l), e = l.lastEffect, e === null ? l.lastEffect = t.next = t : (a = e.next, e.next = t, t.next = a, l.lastEffect = t), t;
  }
  function Nn() {
    return { destroy: void 0, resource: void 0 };
  }
  function rr() {
    return Gt().memoizedState;
  }
  function Dn(t, l, e, a) {
    var u = ol();
    a = a === void 0 ? null : a, tt.flags |= t, u.memoizedState = Ha(
      1 | l,
      Nn(),
      e,
      a
    );
  }
  function Tu(t, l, e, a) {
    var u = Gt();
    a = a === void 0 ? null : a;
    var n = u.memoizedState.inst;
    vt !== null && a !== null && fi(a, vt.memoizedState.deps) ? u.memoizedState = Ha(l, n, e, a) : (tt.flags |= t, u.memoizedState = Ha(
      1 | l,
      n,
      e,
      a
    ));
  }
  function or(t, l) {
    Dn(8390656, 8, t, l);
  }
  function dr(t, l) {
    Tu(2048, 8, t, l);
  }
  function hr(t, l) {
    return Tu(4, 2, t, l);
  }
  function yr(t, l) {
    return Tu(4, 4, t, l);
  }
  function vr(t, l) {
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
  function mr(t, l, e) {
    e = e != null ? e.concat([t]) : null, Tu(4, 4, vr.bind(null, l, t), e);
  }
  function bi() {
  }
  function gr(t, l) {
    var e = Gt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    return l !== null && fi(l, a[1]) ? a[0] : (e.memoizedState = [t, l], t);
  }
  function Sr(t, l) {
    var e = Gt();
    l = l === void 0 ? null : l;
    var a = e.memoizedState;
    if (l !== null && fi(l, a[1]))
      return a[0];
    if (a = t(), ea) {
      Yl(!0);
      try {
        t();
      } finally {
        Yl(!1);
      }
    }
    return e.memoizedState = [a, l], a;
  }
  function xi(t, l, e) {
    return e === void 0 || (Ee & 1073741824) !== 0 ? t.memoizedState = l : (t.memoizedState = e, t = jo(), tt.lanes |= t, _e |= t, e);
  }
  function br(t, l, e, a) {
    return gl(e, l) ? e : _a.current !== null ? (t = xi(t, e, a), gl(t, l) || (Kt = !0), t) : (Ee & 42) === 0 ? (Kt = !0, t.memoizedState = e) : (t = jo(), tt.lanes |= t, _e |= t, l);
  }
  function xr(t, l, e, a, u) {
    var n = C.p;
    C.p = n !== 0 && 8 > n ? n : 8;
    var c = p.T, f = {};
    p.T = f, Ti(t, !1, l, e);
    try {
      var s = u(), m = p.S;
      if (m !== null && m(f, s), s !== null && typeof s == "object" && typeof s.then == "function") {
        var j = Lh(
          s,
          a
        );
        Au(
          t,
          l,
          j,
          jl(t)
        );
      } else
        Au(
          t,
          l,
          a,
          jl(t)
        );
    } catch (A) {
      Au(
        t,
        l,
        { then: function() {
        }, status: "rejected", reason: A },
        jl()
      );
    } finally {
      C.p = n, p.T = c;
    }
  }
  function kh() {
  }
  function pi(t, l, e, a) {
    if (t.tag !== 5) throw Error(r(476));
    var u = pr(t).queue;
    xr(
      t,
      u,
      l,
      V,
      e === null ? kh : function() {
        return jr(t), e(a);
      }
    );
  }
  function pr(t) {
    var l = t.memoizedState;
    if (l !== null) return l;
    l = {
      memoizedState: V,
      baseState: V,
      baseQueue: null,
      queue: {
        pending: null,
        lanes: 0,
        dispatch: null,
        lastRenderedReducer: fe,
        lastRenderedState: V
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
        lastRenderedReducer: fe,
        lastRenderedState: e
      },
      next: null
    }, t.memoizedState = l, t = t.alternate, t !== null && (t.memoizedState = l), l;
  }
  function jr(t) {
    var l = pr(t).next.queue;
    Au(t, l, {}, jl());
  }
  function ji() {
    return tl(Zu);
  }
  function Tr() {
    return Gt().memoizedState;
  }
  function Ar() {
    return Gt().memoizedState;
  }
  function $h(t) {
    for (var l = t.return; l !== null; ) {
      switch (l.tag) {
        case 24:
        case 3:
          var e = jl();
          t = Te(e);
          var a = Ae(l, t, e);
          a !== null && (Tl(a, l, e), Su(a, l, e)), l = { cache: Pc() }, t.payload = l;
          return;
      }
      l = l.return;
    }
  }
  function Wh(t, l, e) {
    var a = jl();
    e = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, On(t) ? zr(l, e) : (e = Zc(t, l, e, a), e !== null && (Tl(e, t, a), Mr(e, l, a)));
  }
  function Er(t, l, e) {
    var a = jl();
    Au(t, l, e, a);
  }
  function Au(t, l, e, a) {
    var u = {
      lane: a,
      revertLane: 0,
      action: e,
      hasEagerState: !1,
      eagerState: null,
      next: null
    };
    if (On(t)) zr(l, u);
    else {
      var n = t.alternate;
      if (t.lanes === 0 && (n === null || n.lanes === 0) && (n = l.lastRenderedReducer, n !== null))
        try {
          var c = l.lastRenderedState, f = n(c, e);
          if (u.hasEagerState = !0, u.eagerState = f, gl(f, c))
            return hn(t, l, u, 0), Tt === null && dn(), !1;
        } catch {
        }
      if (e = Zc(t, l, u, a), e !== null)
        return Tl(e, t, a), Mr(e, l, a), !0;
    }
    return !1;
  }
  function Ti(t, l, e, a) {
    if (a = {
      lane: 2,
      revertLane: lf(),
      action: a,
      hasEagerState: !1,
      eagerState: null,
      next: null
    }, On(t)) {
      if (l) throw Error(r(479));
    } else
      l = Zc(
        t,
        e,
        a,
        2
      ), l !== null && Tl(l, t, 2);
  }
  function On(t) {
    var l = t.alternate;
    return t === tt || l !== null && l === tt;
  }
  function zr(t, l) {
    Ua = An = !0;
    var e = t.pending;
    e === null ? l.next = l : (l.next = e.next, e.next = l), t.pending = l;
  }
  function Mr(t, l, e) {
    if ((e & 4194048) !== 0) {
      var a = l.lanes;
      a &= t.pendingLanes, e |= a, l.lanes = e, Kl(t, e);
    }
  }
  var Rn = {
    readContext: tl,
    use: zn,
    useCallback: Ht,
    useContext: Ht,
    useEffect: Ht,
    useImperativeHandle: Ht,
    useLayoutEffect: Ht,
    useInsertionEffect: Ht,
    useMemo: Ht,
    useReducer: Ht,
    useRef: Ht,
    useState: Ht,
    useDebugValue: Ht,
    useDeferredValue: Ht,
    useTransition: Ht,
    useSyncExternalStore: Ht,
    useId: Ht,
    useHostTransitionStatus: Ht,
    useFormState: Ht,
    useActionState: Ht,
    useOptimistic: Ht,
    useMemoCache: Ht,
    useCacheRefresh: Ht
  }, Nr = {
    readContext: tl,
    use: zn,
    useCallback: function(t, l) {
      return ol().memoizedState = [
        t,
        l === void 0 ? null : l
      ], t;
    },
    useContext: tl,
    useEffect: or,
    useImperativeHandle: function(t, l, e) {
      e = e != null ? e.concat([t]) : null, Dn(
        4194308,
        4,
        vr.bind(null, l, t),
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
      var e = ol();
      l = l === void 0 ? null : l;
      var a = t();
      if (ea) {
        Yl(!0);
        try {
          t();
        } finally {
          Yl(!1);
        }
      }
      return e.memoizedState = [a, l], a;
    },
    useReducer: function(t, l, e) {
      var a = ol();
      if (e !== void 0) {
        var u = e(l);
        if (ea) {
          Yl(!0);
          try {
            e(l);
          } finally {
            Yl(!1);
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
        tt,
        t
      ), [a.memoizedState, t];
    },
    useRef: function(t) {
      var l = ol();
      return t = { current: t }, l.memoizedState = t;
    },
    useState: function(t) {
      t = gi(t);
      var l = t.queue, e = Er.bind(null, tt, l);
      return l.dispatch = e, [t.memoizedState, e];
    },
    useDebugValue: bi,
    useDeferredValue: function(t, l) {
      var e = ol();
      return xi(e, t, l);
    },
    useTransition: function() {
      var t = gi(!1);
      return t = xr.bind(
        null,
        tt,
        t.queue,
        !0,
        !1
      ), ol().memoizedState = t, [!1, t];
    },
    useSyncExternalStore: function(t, l, e) {
      var a = tt, u = ol();
      if (dt) {
        if (e === void 0)
          throw Error(r(407));
        e = e();
      } else {
        if (e = l(), Tt === null)
          throw Error(r(349));
        (ct & 124) !== 0 || $s(a, l, e);
      }
      u.memoizedState = e;
      var n = { value: e, getSnapshot: l };
      return u.queue = n, or(Fs.bind(null, a, n, t), [
        t
      ]), a.flags |= 2048, Ha(
        9,
        Nn(),
        Ws.bind(
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
      var t = ol(), l = Tt.identifierPrefix;
      if (dt) {
        var e = ne, a = ue;
        e = (a & ~(1 << 32 - ul(a) - 1)).toString(32) + e, l = "«" + l + "R" + e, e = En++, 0 < e && (l += "H" + e.toString(32)), l += "»";
      } else
        e = Vh++, l = "«" + l + "r" + e.toString(32) + "»";
      return t.memoizedState = l;
    },
    useHostTransitionStatus: ji,
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
      return l.queue = e, l = Ti.bind(
        null,
        tt,
        !0,
        e
      ), e.dispatch = l, [t, l];
    },
    useMemoCache: yi,
    useCacheRefresh: function() {
      return ol().memoizedState = $h.bind(
        null,
        tt
      );
    }
  }, Dr = {
    readContext: tl,
    use: zn,
    useCallback: gr,
    useContext: tl,
    useEffect: dr,
    useImperativeHandle: mr,
    useInsertionEffect: hr,
    useLayoutEffect: yr,
    useMemo: Sr,
    useReducer: Mn,
    useRef: rr,
    useState: function() {
      return Mn(fe);
    },
    useDebugValue: bi,
    useDeferredValue: function(t, l) {
      var e = Gt();
      return br(
        e,
        vt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = Mn(fe)[0], l = Gt().memoizedState;
      return [
        typeof t == "boolean" ? t : ju(t),
        l
      ];
    },
    useSyncExternalStore: ks,
    useId: Tr,
    useHostTransitionStatus: ji,
    useFormState: ir,
    useActionState: ir,
    useOptimistic: function(t, l) {
      var e = Gt();
      return tr(e, vt, t, l);
    },
    useMemoCache: yi,
    useCacheRefresh: Ar
  }, Fh = {
    readContext: tl,
    use: zn,
    useCallback: gr,
    useContext: tl,
    useEffect: dr,
    useImperativeHandle: mr,
    useInsertionEffect: hr,
    useLayoutEffect: yr,
    useMemo: Sr,
    useReducer: mi,
    useRef: rr,
    useState: function() {
      return mi(fe);
    },
    useDebugValue: bi,
    useDeferredValue: function(t, l) {
      var e = Gt();
      return vt === null ? xi(e, t, l) : br(
        e,
        vt.memoizedState,
        t,
        l
      );
    },
    useTransition: function() {
      var t = mi(fe)[0], l = Gt().memoizedState;
      return [
        typeof t == "boolean" ? t : ju(t),
        l
      ];
    },
    useSyncExternalStore: ks,
    useId: Tr,
    useHostTransitionStatus: ji,
    useFormState: sr,
    useActionState: sr,
    useOptimistic: function(t, l) {
      var e = Gt();
      return vt !== null ? tr(e, vt, t, l) : (e.baseState = t, [t, e.queue.dispatch]);
    },
    useMemoCache: yi,
    useCacheRefresh: Ar
  }, qa = null, Eu = 0;
  function _n(t) {
    var l = Eu;
    return Eu += 1, qa === null && (qa = []), Xs(qa, t, l);
  }
  function zu(t, l) {
    l = l.props.ref, t.ref = l !== void 0 ? l : null;
  }
  function Un(t, l) {
    throw l.$$typeof === X ? Error(r(525)) : (t = Object.prototype.toString.call(l), Error(
      r(
        31,
        t === "[object Object]" ? "object with keys {" + Object.keys(l).join(", ") + "}" : t
      )
    ));
  }
  function Or(t) {
    var l = t._init;
    return l(t._payload);
  }
  function Rr(t) {
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
      return y = ae(y, h), y.index = 0, y.sibling = null, y;
    }
    function n(y, h, v) {
      return y.index = v, t ? (v = y.alternate, v !== null ? (v = v.index, v < h ? (y.flags |= 67108866, h) : v) : (y.flags |= 67108866, h)) : (y.flags |= 1048576, h);
    }
    function c(y) {
      return t && y.alternate === null && (y.flags |= 67108866), y;
    }
    function f(y, h, v, T) {
      return h === null || h.tag !== 6 ? (h = Vc(v, y.mode, T), h.return = y, h) : (h = u(h, v), h.return = y, h);
    }
    function s(y, h, v, T) {
      var G = v.type;
      return G === q ? j(
        y,
        h,
        v.props.children,
        T,
        v.key
      ) : h !== null && (h.elementType === G || typeof G == "object" && G !== null && G.$$typeof === zt && Or(G) === h.type) ? (h = u(h, v.props), zu(h, v), h.return = y, h) : (h = vn(
        v.type,
        v.key,
        v.props,
        null,
        y.mode,
        T
      ), zu(h, v), h.return = y, h);
    }
    function m(y, h, v, T) {
      return h === null || h.tag !== 4 || h.stateNode.containerInfo !== v.containerInfo || h.stateNode.implementation !== v.implementation ? (h = wc(v, y.mode, T), h.return = y, h) : (h = u(h, v.children || []), h.return = y, h);
    }
    function j(y, h, v, T, G) {
      return h === null || h.tag !== 7 ? (h = ke(
        v,
        y.mode,
        T,
        G
      ), h.return = y, h) : (h = u(h, v), h.return = y, h);
    }
    function A(y, h, v) {
      if (typeof h == "string" && h !== "" || typeof h == "number" || typeof h == "bigint")
        return h = Vc(
          "" + h,
          y.mode,
          v
        ), h.return = y, h;
      if (typeof h == "object" && h !== null) {
        switch (h.$$typeof) {
          case W:
            return v = vn(
              h.type,
              h.key,
              h.props,
              null,
              y.mode,
              v
            ), zu(v, h), v.return = y, v;
          case it:
            return h = wc(
              h,
              y.mode,
              v
            ), h.return = y, h;
          case zt:
            var T = h._init;
            return h = T(h._payload), A(y, h, v);
        }
        if (Bt(h) || Ot(h))
          return h = ke(
            h,
            y.mode,
            v,
            null
          ), h.return = y, h;
        if (typeof h.then == "function")
          return A(y, _n(h), v);
        if (h.$$typeof === Et)
          return A(
            y,
            bn(y, h),
            v
          );
        Un(y, h);
      }
      return null;
    }
    function g(y, h, v, T) {
      var G = h !== null ? h.key : null;
      if (typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint")
        return G !== null ? null : f(y, h, "" + v, T);
      if (typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case W:
            return v.key === G ? s(y, h, v, T) : null;
          case it:
            return v.key === G ? m(y, h, v, T) : null;
          case zt:
            return G = v._init, v = G(v._payload), g(y, h, v, T);
        }
        if (Bt(v) || Ot(v))
          return G !== null ? null : j(y, h, v, T, null);
        if (typeof v.then == "function")
          return g(
            y,
            h,
            _n(v),
            T
          );
        if (v.$$typeof === Et)
          return g(
            y,
            h,
            bn(y, v),
            T
          );
        Un(y, v);
      }
      return null;
    }
    function S(y, h, v, T, G) {
      if (typeof T == "string" && T !== "" || typeof T == "number" || typeof T == "bigint")
        return y = y.get(v) || null, f(h, y, "" + T, G);
      if (typeof T == "object" && T !== null) {
        switch (T.$$typeof) {
          case W:
            return y = y.get(
              T.key === null ? v : T.key
            ) || null, s(h, y, T, G);
          case it:
            return y = y.get(
              T.key === null ? v : T.key
            ) || null, m(h, y, T, G);
          case zt:
            var lt = T._init;
            return T = lt(T._payload), S(
              y,
              h,
              v,
              T,
              G
            );
        }
        if (Bt(T) || Ot(T))
          return y = y.get(v) || null, j(h, y, T, G, null);
        if (typeof T.then == "function")
          return S(
            y,
            h,
            v,
            _n(T),
            G
          );
        if (T.$$typeof === Et)
          return S(
            y,
            h,
            v,
            bn(h, T),
            G
          );
        Un(h, T);
      }
      return null;
    }
    function k(y, h, v, T) {
      for (var G = null, lt = null, L = h, J = h = 0, kt = null; L !== null && J < v.length; J++) {
        L.index > J ? (kt = L, L = null) : kt = L.sibling;
        var rt = g(
          y,
          L,
          v[J],
          T
        );
        if (rt === null) {
          L === null && (L = kt);
          break;
        }
        t && L && rt.alternate === null && l(y, L), h = n(rt, h, J), lt === null ? G = rt : lt.sibling = rt, lt = rt, L = kt;
      }
      if (J === v.length)
        return e(y, L), dt && We(y, J), G;
      if (L === null) {
        for (; J < v.length; J++)
          L = A(y, v[J], T), L !== null && (h = n(
            L,
            h,
            J
          ), lt === null ? G = L : lt.sibling = L, lt = L);
        return dt && We(y, J), G;
      }
      for (L = a(L); J < v.length; J++)
        kt = S(
          L,
          y,
          J,
          v[J],
          T
        ), kt !== null && (t && kt.alternate !== null && L.delete(
          kt.key === null ? J : kt.key
        ), h = n(
          kt,
          h,
          J
        ), lt === null ? G = kt : lt.sibling = kt, lt = kt);
      return t && L.forEach(function(Qe) {
        return l(y, Qe);
      }), dt && We(y, J), G;
    }
    function K(y, h, v, T) {
      if (v == null) throw Error(r(151));
      for (var G = null, lt = null, L = h, J = h = 0, kt = null, rt = v.next(); L !== null && !rt.done; J++, rt = v.next()) {
        L.index > J ? (kt = L, L = null) : kt = L.sibling;
        var Qe = g(y, L, rt.value, T);
        if (Qe === null) {
          L === null && (L = kt);
          break;
        }
        t && L && Qe.alternate === null && l(y, L), h = n(Qe, h, J), lt === null ? G = Qe : lt.sibling = Qe, lt = Qe, L = kt;
      }
      if (rt.done)
        return e(y, L), dt && We(y, J), G;
      if (L === null) {
        for (; !rt.done; J++, rt = v.next())
          rt = A(y, rt.value, T), rt !== null && (h = n(rt, h, J), lt === null ? G = rt : lt.sibling = rt, lt = rt);
        return dt && We(y, J), G;
      }
      for (L = a(L); !rt.done; J++, rt = v.next())
        rt = S(L, y, J, rt.value, T), rt !== null && (t && rt.alternate !== null && L.delete(rt.key === null ? J : rt.key), h = n(rt, h, J), lt === null ? G = rt : lt.sibling = rt, lt = rt);
      return t && L.forEach(function(P0) {
        return l(y, P0);
      }), dt && We(y, J), G;
    }
    function gt(y, h, v, T) {
      if (typeof v == "object" && v !== null && v.type === q && v.key === null && (v = v.props.children), typeof v == "object" && v !== null) {
        switch (v.$$typeof) {
          case W:
            t: {
              for (var G = v.key; h !== null; ) {
                if (h.key === G) {
                  if (G = v.type, G === q) {
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
                  } else if (h.elementType === G || typeof G == "object" && G !== null && G.$$typeof === zt && Or(G) === h.type) {
                    e(
                      y,
                      h.sibling
                    ), T = u(h, v.props), zu(T, v), T.return = y, y = T;
                    break t;
                  }
                  e(y, h);
                  break;
                } else l(y, h);
                h = h.sibling;
              }
              v.type === q ? (T = ke(
                v.props.children,
                y.mode,
                T,
                v.key
              ), T.return = y, y = T) : (T = vn(
                v.type,
                v.key,
                v.props,
                null,
                y.mode,
                T
              ), zu(T, v), T.return = y, y = T);
            }
            return c(y);
          case it:
            t: {
              for (G = v.key; h !== null; ) {
                if (h.key === G)
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
              T = wc(v, y.mode, T), T.return = y, y = T;
            }
            return c(y);
          case zt:
            return G = v._init, v = G(v._payload), gt(
              y,
              h,
              v,
              T
            );
        }
        if (Bt(v))
          return k(
            y,
            h,
            v,
            T
          );
        if (Ot(v)) {
          if (G = Ot(v), typeof G != "function") throw Error(r(150));
          return v = G.call(v), K(
            y,
            h,
            v,
            T
          );
        }
        if (typeof v.then == "function")
          return gt(
            y,
            h,
            _n(v),
            T
          );
        if (v.$$typeof === Et)
          return gt(
            y,
            h,
            bn(y, v),
            T
          );
        Un(y, v);
      }
      return typeof v == "string" && v !== "" || typeof v == "number" || typeof v == "bigint" ? (v = "" + v, h !== null && h.tag === 6 ? (e(y, h.sibling), T = u(h, v), T.return = y, y = T) : (e(y, h), T = Vc(v, y.mode, T), T.return = y, y = T), c(y)) : e(y, h);
    }
    return function(y, h, v, T) {
      try {
        Eu = 0;
        var G = gt(
          y,
          h,
          v,
          T
        );
        return qa = null, G;
      } catch (L) {
        if (L === mu || L === pn) throw L;
        var lt = Sl(29, L, null, y.mode);
        return lt.lanes = T, lt.return = y, lt;
      }
    };
  }
  var Ba = Rr(!0), _r = Rr(!1), Cl = M(null), $l = null;
  function ze(t) {
    var l = t.alternate;
    H(Qt, Qt.current & 1), H(Cl, t), $l === null && (l === null || _a.current !== null || l.memoizedState !== null) && ($l = t);
  }
  function Ur(t) {
    if (t.tag === 22) {
      if (H(Qt, Qt.current), H(Cl, t), $l === null) {
        var l = t.alternate;
        l !== null && l.memoizedState !== null && ($l = t);
      }
    } else Me();
  }
  function Me() {
    H(Qt, Qt.current), H(Cl, Cl.current);
  }
  function se(t) {
    Y(Cl), $l === t && ($l = null), Y(Qt);
  }
  var Qt = M(0);
  function Cn(t) {
    for (var l = t; l !== null; ) {
      if (l.tag === 13) {
        var e = l.memoizedState;
        if (e !== null && (e = e.dehydrated, e === null || e.data === "$?" || yf(e)))
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
  function Ai(t, l, e, a) {
    l = t.memoizedState, e = e(a, l), e = e == null ? l : B({}, l, e), t.memoizedState = e, t.lanes === 0 && (t.updateQueue.baseState = e);
  }
  var Ei = {
    enqueueSetState: function(t, l, e) {
      t = t._reactInternals;
      var a = jl(), u = Te(a);
      u.payload = l, e != null && (u.callback = e), l = Ae(t, u, a), l !== null && (Tl(l, t, a), Su(l, t, a));
    },
    enqueueReplaceState: function(t, l, e) {
      t = t._reactInternals;
      var a = jl(), u = Te(a);
      u.tag = 1, u.payload = l, e != null && (u.callback = e), l = Ae(t, u, a), l !== null && (Tl(l, t, a), Su(l, t, a));
    },
    enqueueForceUpdate: function(t, l) {
      t = t._reactInternals;
      var e = jl(), a = Te(e);
      a.tag = 2, l != null && (a.callback = l), l = Ae(t, a, e), l !== null && (Tl(l, t, e), Su(l, t, e));
    }
  };
  function Cr(t, l, e, a, u, n, c) {
    return t = t.stateNode, typeof t.shouldComponentUpdate == "function" ? t.shouldComponentUpdate(a, n, c) : l.prototype && l.prototype.isPureReactComponent ? !fu(e, a) || !fu(u, n) : !0;
  }
  function Hr(t, l, e, a) {
    t = l.state, typeof l.componentWillReceiveProps == "function" && l.componentWillReceiveProps(e, a), typeof l.UNSAFE_componentWillReceiveProps == "function" && l.UNSAFE_componentWillReceiveProps(e, a), l.state !== t && Ei.enqueueReplaceState(l, l.state, null);
  }
  function aa(t, l) {
    var e = l;
    if ("ref" in l) {
      e = {};
      for (var a in l)
        a !== "ref" && (e[a] = l[a]);
    }
    if (t = t.defaultProps) {
      e === l && (e = B({}, e));
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
  function qr(t) {
    Hn(t);
  }
  function Br(t) {
    console.error(t);
  }
  function Yr(t) {
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
  function Gr(t, l, e) {
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
  function zi(t, l, e) {
    return e = Te(e), e.tag = 3, e.payload = { element: null }, e.callback = function() {
      qn(t, l);
    }, e;
  }
  function Xr(t) {
    return t = Te(t), t.tag = 3, t;
  }
  function Qr(t, l, e, a) {
    var u = e.type.getDerivedStateFromError;
    if (typeof u == "function") {
      var n = a.value;
      t.payload = function() {
        return u(n);
      }, t.callback = function() {
        Gr(l, e, a);
      };
    }
    var c = e.stateNode;
    c !== null && typeof c.componentDidCatch == "function" && (t.callback = function() {
      Gr(l, e, a), typeof u != "function" && (Ue === null ? Ue = /* @__PURE__ */ new Set([this]) : Ue.add(this));
      var f = a.stack;
      this.componentDidCatch(a.value, {
        componentStack: f !== null ? f : ""
      });
    });
  }
  function Ph(t, l, e, a, u) {
    if (e.flags |= 32768, a !== null && typeof a == "object" && typeof a.then == "function") {
      if (l = e.alternate, l !== null && hu(
        l,
        e,
        u,
        !0
      ), e = Cl.current, e !== null) {
        switch (e.tag) {
          case 13:
            return $l === null ? Wi() : e.alternate === null && Ut === 0 && (Ut = 3), e.flags &= -257, e.flags |= 65536, e.lanes = u, a === li ? e.flags |= 16384 : (l = e.updateQueue, l === null ? e.updateQueue = /* @__PURE__ */ new Set([a]) : l.add(a), Pi(t, a, u)), !1;
          case 22:
            return e.flags |= 65536, a === li ? e.flags |= 16384 : (l = e.updateQueue, l === null ? (l = {
              transitions: null,
              markerInstances: null,
              retryQueue: /* @__PURE__ */ new Set([a])
            }, e.updateQueue = l) : (e = l.retryQueue, e === null ? l.retryQueue = /* @__PURE__ */ new Set([a]) : e.add(a)), Pi(t, a, u)), !1;
        }
        throw Error(r(435, e.tag));
      }
      return Pi(t, a, u), Wi(), !1;
    }
    if (dt)
      return l = Cl.current, l !== null ? ((l.flags & 65536) === 0 && (l.flags |= 256), l.flags |= 65536, l.lanes = u, a !== kc && (t = Error(r(422), { cause: a }), du(Ol(t, e)))) : (a !== kc && (l = Error(r(423), {
        cause: a
      }), du(
        Ol(l, e)
      )), t = t.current.alternate, t.flags |= 65536, u &= -u, t.lanes |= u, a = Ol(a, e), u = zi(
        t.stateNode,
        a,
        u
      ), ui(t, u), Ut !== 4 && (Ut = 2)), !1;
    var n = Error(r(520), { cause: a });
    if (n = Ol(n, e), Uu === null ? Uu = [n] : Uu.push(n), Ut !== 4 && (Ut = 2), l === null) return !0;
    a = Ol(a, e), e = l;
    do {
      switch (e.tag) {
        case 3:
          return e.flags |= 65536, t = u & -u, e.lanes |= t, t = zi(e.stateNode, a, t), ui(e, t), !1;
        case 1:
          if (l = e.type, n = e.stateNode, (e.flags & 128) === 0 && (typeof l.getDerivedStateFromError == "function" || n !== null && typeof n.componentDidCatch == "function" && (Ue === null || !Ue.has(n))))
            return e.flags |= 65536, u &= -u, e.lanes |= u, u = Xr(u), Qr(
              u,
              t,
              e,
              a
            ), ui(e, u), !1;
      }
      e = e.return;
    } while (e !== null);
    return !1;
  }
  var Zr = Error(r(461)), Kt = !1;
  function Wt(t, l, e, a) {
    l.child = t === null ? _r(l, null, e, a) : Ba(
      l,
      t.child,
      e,
      a
    );
  }
  function Lr(t, l, e, a, u) {
    e = e.render;
    var n = l.ref;
    if ("ref" in a) {
      var c = {};
      for (var f in a)
        f !== "ref" && (c[f] = a[f]);
    } else c = a;
    return ta(l), a = si(
      t,
      l,
      e,
      c,
      n,
      u
    ), f = ri(), t !== null && !Kt ? (oi(t, l, u), re(t, l, u)) : (dt && f && Kc(l), l.flags |= 1, Wt(t, l, a, u), l.child);
  }
  function Vr(t, l, e, a, u) {
    if (t === null) {
      var n = e.type;
      return typeof n == "function" && !Lc(n) && n.defaultProps === void 0 && e.compare === null ? (l.tag = 15, l.type = n, wr(
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
    if (n = t.child, !Ci(t, u)) {
      var c = n.memoizedProps;
      if (e = e.compare, e = e !== null ? e : fu, e(c, a) && t.ref === l.ref)
        return re(t, l, u);
    }
    return l.flags |= 1, t = ae(n, a), t.ref = l.ref, t.return = l, l.child = t;
  }
  function wr(t, l, e, a, u) {
    if (t !== null) {
      var n = t.memoizedProps;
      if (fu(n, a) && t.ref === l.ref)
        if (Kt = !1, l.pendingProps = a = n, Ci(t, u))
          (t.flags & 131072) !== 0 && (Kt = !0);
        else
          return l.lanes = t.lanes, re(t, l, u);
    }
    return Mi(
      t,
      l,
      e,
      a,
      u
    );
  }
  function Kr(t, l, e) {
    var a = l.pendingProps, u = a.children, n = t !== null ? t.memoizedState : null;
    if (a.mode === "hidden") {
      if ((l.flags & 128) !== 0) {
        if (a = n !== null ? n.baseLanes | e : e, t !== null) {
          for (u = l.child = t.child, n = 0; u !== null; )
            n = n | u.lanes | u.childLanes, u = u.sibling;
          l.childLanes = n & ~a;
        } else l.childLanes = 0, l.child = null;
        return Jr(
          t,
          l,
          a,
          e
        );
      }
      if ((e & 536870912) !== 0)
        l.memoizedState = { baseLanes: 0, cachePool: null }, t !== null && xn(
          l,
          n !== null ? n.cachePool : null
        ), n !== null ? ws(l, n) : ci(), Ur(l);
      else
        return l.lanes = l.childLanes = 536870912, Jr(
          t,
          l,
          n !== null ? n.baseLanes | e : e,
          e
        );
    } else
      n !== null ? (xn(l, n.cachePool), ws(l, n), Me(), l.memoizedState = null) : (t !== null && xn(l, null), ci(), Me());
    return Wt(t, l, u, e), l.child;
  }
  function Jr(t, l, e, a) {
    var u = ti();
    return u = u === null ? null : { parent: Xt._currentValue, pool: u }, l.memoizedState = {
      baseLanes: e,
      cachePool: u
    }, t !== null && xn(l, null), ci(), Ur(l), t !== null && hu(t, l, a, !0), null;
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
  function Mi(t, l, e, a, u) {
    return ta(l), e = si(
      t,
      l,
      e,
      a,
      void 0,
      u
    ), a = ri(), t !== null && !Kt ? (oi(t, l, u), re(t, l, u)) : (dt && a && Kc(l), l.flags |= 1, Wt(t, l, e, u), l.child);
  }
  function kr(t, l, e, a, u, n) {
    return ta(l), l.updateQueue = null, e = Js(
      l,
      a,
      e,
      u
    ), Ks(t), a = ri(), t !== null && !Kt ? (oi(t, l, n), re(t, l, n)) : (dt && a && Kc(l), l.flags |= 1, Wt(t, l, e, n), l.child);
  }
  function $r(t, l, e, a, u) {
    if (ta(l), l.stateNode === null) {
      var n = Ma, c = e.contextType;
      typeof c == "object" && c !== null && (n = tl(c)), n = new e(a, n), l.memoizedState = n.state !== null && n.state !== void 0 ? n.state : null, n.updater = Ei, l.stateNode = n, n._reactInternals = l, n = l.stateNode, n.props = a, n.state = l.memoizedState, n.refs = {}, ei(l), c = e.contextType, n.context = typeof c == "object" && c !== null ? tl(c) : Ma, n.state = l.memoizedState, c = e.getDerivedStateFromProps, typeof c == "function" && (Ai(
        l,
        e,
        c,
        a
      ), n.state = l.memoizedState), typeof e.getDerivedStateFromProps == "function" || typeof n.getSnapshotBeforeUpdate == "function" || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (c = n.state, typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount(), c !== n.state && Ei.enqueueReplaceState(n, n.state, null), xu(l, a, n, u), bu(), n.state = l.memoizedState), typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !0;
    } else if (t === null) {
      n = l.stateNode;
      var f = l.memoizedProps, s = aa(e, f);
      n.props = s;
      var m = n.context, j = e.contextType;
      c = Ma, typeof j == "object" && j !== null && (c = tl(j));
      var A = e.getDerivedStateFromProps;
      j = typeof A == "function" || typeof n.getSnapshotBeforeUpdate == "function", f = l.pendingProps !== f, j || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (f || m !== c) && Hr(
        l,
        n,
        a,
        c
      ), je = !1;
      var g = l.memoizedState;
      n.state = g, xu(l, a, n, u), bu(), m = l.memoizedState, f || g !== m || je ? (typeof A == "function" && (Ai(
        l,
        e,
        A,
        a
      ), m = l.memoizedState), (s = je || Cr(
        l,
        e,
        s,
        a,
        g,
        m,
        c
      )) ? (j || typeof n.UNSAFE_componentWillMount != "function" && typeof n.componentWillMount != "function" || (typeof n.componentWillMount == "function" && n.componentWillMount(), typeof n.UNSAFE_componentWillMount == "function" && n.UNSAFE_componentWillMount()), typeof n.componentDidMount == "function" && (l.flags |= 4194308)) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), l.memoizedProps = a, l.memoizedState = m), n.props = a, n.state = m, n.context = c, a = s) : (typeof n.componentDidMount == "function" && (l.flags |= 4194308), a = !1);
    } else {
      n = l.stateNode, ai(t, l), c = l.memoizedProps, j = aa(e, c), n.props = j, A = l.pendingProps, g = n.context, m = e.contextType, s = Ma, typeof m == "object" && m !== null && (s = tl(m)), f = e.getDerivedStateFromProps, (m = typeof f == "function" || typeof n.getSnapshotBeforeUpdate == "function") || typeof n.UNSAFE_componentWillReceiveProps != "function" && typeof n.componentWillReceiveProps != "function" || (c !== A || g !== s) && Hr(
        l,
        n,
        a,
        s
      ), je = !1, g = l.memoizedState, n.state = g, xu(l, a, n, u), bu();
      var S = l.memoizedState;
      c !== A || g !== S || je || t !== null && t.dependencies !== null && Sn(t.dependencies) ? (typeof f == "function" && (Ai(
        l,
        e,
        f,
        a
      ), S = l.memoizedState), (j = je || Cr(
        l,
        e,
        j,
        a,
        g,
        S,
        s
      ) || t !== null && t.dependencies !== null && Sn(t.dependencies)) ? (m || typeof n.UNSAFE_componentWillUpdate != "function" && typeof n.componentWillUpdate != "function" || (typeof n.componentWillUpdate == "function" && n.componentWillUpdate(a, S, s), typeof n.UNSAFE_componentWillUpdate == "function" && n.UNSAFE_componentWillUpdate(
        a,
        S,
        s
      )), typeof n.componentDidUpdate == "function" && (l.flags |= 4), typeof n.getSnapshotBeforeUpdate == "function" && (l.flags |= 1024)) : (typeof n.componentDidUpdate != "function" || c === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || c === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), l.memoizedProps = a, l.memoizedState = S), n.props = a, n.state = S, n.context = s, a = j) : (typeof n.componentDidUpdate != "function" || c === t.memoizedProps && g === t.memoizedState || (l.flags |= 4), typeof n.getSnapshotBeforeUpdate != "function" || c === t.memoizedProps && g === t.memoizedState || (l.flags |= 1024), a = !1);
    }
    return n = a, Bn(t, l), a = (l.flags & 128) !== 0, n || a ? (n = l.stateNode, e = a && typeof e.getDerivedStateFromError != "function" ? null : n.render(), l.flags |= 1, t !== null && a ? (l.child = Ba(
      l,
      t.child,
      null,
      u
    ), l.child = Ba(
      l,
      null,
      e,
      u
    )) : Wt(t, l, e, u), l.memoizedState = n.state, t = l.child) : t = re(
      t,
      l,
      u
    ), t;
  }
  function Wr(t, l, e, a) {
    return ou(), l.flags |= 256, Wt(t, l, e, a), l.child;
  }
  var Ni = {
    dehydrated: null,
    treeContext: null,
    retryLane: 0,
    hydrationErrors: null
  };
  function Di(t) {
    return { baseLanes: t, cachePool: Bs() };
  }
  function Oi(t, l, e) {
    return t = t !== null ? t.childLanes & ~e : 0, l && (t |= Hl), t;
  }
  function Fr(t, l, e) {
    var a = l.pendingProps, u = !1, n = (l.flags & 128) !== 0, c;
    if ((c = n) || (c = t !== null && t.memoizedState === null ? !1 : (Qt.current & 2) !== 0), c && (u = !0, l.flags &= -129), c = (l.flags & 32) !== 0, l.flags &= -33, t === null) {
      if (dt) {
        if (u ? ze(l) : Me(), dt) {
          var f = _t, s;
          if (s = f) {
            t: {
              for (s = f, f = kl; s.nodeType !== 8; ) {
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
              treeContext: $e !== null ? { id: ue, overflow: ne } : null,
              retryLane: 536870912,
              hydrationErrors: null
            }, s = Sl(
              18,
              null,
              null,
              0
            ), s.stateNode = f, s.return = l, l.child = s, nl = l, _t = null, s = !0) : s = !1;
          }
          s || Pe(l);
        }
        if (f = l.memoizedState, f !== null && (f = f.dehydrated, f !== null))
          return yf(f) ? l.lanes = 32 : l.lanes = 536870912, null;
        se(l);
      }
      return f = a.children, a = a.fallback, u ? (Me(), u = l.mode, f = Yn(
        { mode: "hidden", children: f },
        u
      ), a = ke(
        a,
        u,
        e,
        null
      ), f.return = l, a.return = l, f.sibling = a, l.child = f, u = l.child, u.memoizedState = Di(e), u.childLanes = Oi(
        t,
        c,
        e
      ), l.memoizedState = Ni, a) : (ze(l), Ri(l, f));
    }
    if (s = t.memoizedState, s !== null && (f = s.dehydrated, f !== null)) {
      if (n)
        l.flags & 256 ? (ze(l), l.flags &= -257, l = _i(
          t,
          l,
          e
        )) : l.memoizedState !== null ? (Me(), l.child = t.child, l.flags |= 128, l = null) : (Me(), u = a.fallback, f = l.mode, a = Yn(
          { mode: "visible", children: a.children },
          f
        ), u = ke(
          u,
          f,
          e,
          null
        ), u.flags |= 2, a.return = l, u.return = l, a.sibling = u, l.child = a, Ba(
          l,
          t.child,
          null,
          e
        ), a = l.child, a.memoizedState = Di(e), a.childLanes = Oi(
          t,
          c,
          e
        ), l.memoizedState = Ni, l = u);
      else if (ze(l), yf(f)) {
        if (c = f.nextSibling && f.nextSibling.dataset, c) var m = c.dgst;
        c = m, a = Error(r(419)), a.stack = "", a.digest = c, du({ value: a, source: null, stack: null }), l = _i(
          t,
          l,
          e
        );
      } else if (Kt || hu(t, l, e, !1), c = (e & t.childLanes) !== 0, Kt || c) {
        if (c = Tt, c !== null && (a = e & -e, a = (a & 42) !== 0 ? 1 : Jl(a), a = (a & (c.suspendedLanes | e)) !== 0 ? 0 : a, a !== 0 && a !== s.retryLane))
          throw s.retryLane = a, za(t, a), Tl(c, t, a), Zr;
        f.data === "$?" || Wi(), l = _i(
          t,
          l,
          e
        );
      } else
        f.data === "$?" ? (l.flags |= 192, l.child = t.child, l = null) : (t = s.treeContext, _t = Ql(
          f.nextSibling
        ), nl = l, dt = !0, Fe = null, kl = !1, t !== null && (_l[Ul++] = ue, _l[Ul++] = ne, _l[Ul++] = $e, ue = t.id, ne = t.overflow, $e = l), l = Ri(
          l,
          a.children
        ), l.flags |= 4096);
      return l;
    }
    return u ? (Me(), u = a.fallback, f = l.mode, s = t.child, m = s.sibling, a = ae(s, {
      mode: "hidden",
      children: a.children
    }), a.subtreeFlags = s.subtreeFlags & 65011712, m !== null ? u = ae(m, u) : (u = ke(
      u,
      f,
      e,
      null
    ), u.flags |= 2), u.return = l, a.return = l, a.sibling = u, l.child = a, a = u, u = l.child, f = t.child.memoizedState, f === null ? f = Di(e) : (s = f.cachePool, s !== null ? (m = Xt._currentValue, s = s.parent !== m ? { parent: m, pool: m } : s) : s = Bs(), f = {
      baseLanes: f.baseLanes | e,
      cachePool: s
    }), u.memoizedState = f, u.childLanes = Oi(
      t,
      c,
      e
    ), l.memoizedState = Ni, a) : (ze(l), e = t.child, t = e.sibling, e = ae(e, {
      mode: "visible",
      children: a.children
    }), e.return = l, e.sibling = null, t !== null && (c = l.deletions, c === null ? (l.deletions = [t], l.flags |= 16) : c.push(t)), l.child = e, l.memoizedState = null, e);
  }
  function Ri(t, l) {
    return l = Yn(
      { mode: "visible", children: l },
      t.mode
    ), l.return = t, t.child = l;
  }
  function Yn(t, l) {
    return t = Sl(22, t, null, l), t.lanes = 0, t.stateNode = {
      _visibility: 1,
      _pendingMarkers: null,
      _retryCache: null,
      _transitions: null
    }, t;
  }
  function _i(t, l, e) {
    return Ba(l, t.child, null, e), t = Ri(
      l,
      l.pendingProps.children
    ), t.flags |= 2, l.memoizedState = null, t;
  }
  function Pr(t, l, e) {
    t.lanes |= l;
    var a = t.alternate;
    a !== null && (a.lanes |= l), Wc(t.return, l, e);
  }
  function Ui(t, l, e, a, u) {
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
  function Ir(t, l, e) {
    var a = l.pendingProps, u = a.revealOrder, n = a.tail;
    if (Wt(t, l, a.children, e), a = Qt.current, (a & 2) !== 0)
      a = a & 1 | 2, l.flags |= 128;
    else {
      if (t !== null && (t.flags & 128) !== 0)
        t: for (t = l.child; t !== null; ) {
          if (t.tag === 13)
            t.memoizedState !== null && Pr(t, e, l);
          else if (t.tag === 19)
            Pr(t, e, l);
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
    switch (H(Qt, a), u) {
      case "forwards":
        for (e = l.child, u = null; e !== null; )
          t = e.alternate, t !== null && Cn(t) === null && (u = e), e = e.sibling;
        e = u, e === null ? (u = l.child, l.child = null) : (u = e.sibling, e.sibling = null), Ui(
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
        Ui(
          l,
          !0,
          e,
          null,
          n
        );
        break;
      case "together":
        Ui(l, !1, null, null, void 0);
        break;
      default:
        l.memoizedState = null;
    }
    return l.child;
  }
  function re(t, l, e) {
    if (t !== null && (l.dependencies = t.dependencies), _e |= l.lanes, (e & l.childLanes) === 0)
      if (t !== null) {
        if (hu(
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
      for (t = l.child, e = ae(t, t.pendingProps), l.child = e, e.return = l; t.sibling !== null; )
        t = t.sibling, e = e.sibling = ae(t, t.pendingProps), e.return = l;
      e.sibling = null;
    }
    return l.child;
  }
  function Ci(t, l) {
    return (t.lanes & l) !== 0 ? !0 : (t = t.dependencies, !!(t !== null && Sn(t)));
  }
  function Ih(t, l, e) {
    switch (l.tag) {
      case 3:
        pt(l, l.stateNode.containerInfo), pe(l, Xt, t.memoizedState.cache), ou();
        break;
      case 27:
      case 5:
        Bl(l);
        break;
      case 4:
        pt(l, l.stateNode.containerInfo);
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
          return a.dehydrated !== null ? (ze(l), l.flags |= 128, null) : (e & l.child.childLanes) !== 0 ? Fr(t, l, e) : (ze(l), t = re(
            t,
            l,
            e
          ), t !== null ? t.sibling : null);
        ze(l);
        break;
      case 19:
        var u = (t.flags & 128) !== 0;
        if (a = (e & l.childLanes) !== 0, a || (hu(
          t,
          l,
          e,
          !1
        ), a = (e & l.childLanes) !== 0), u) {
          if (a)
            return Ir(
              t,
              l,
              e
            );
          l.flags |= 128;
        }
        if (u = l.memoizedState, u !== null && (u.rendering = null, u.tail = null, u.lastEffect = null), H(Qt, Qt.current), a) break;
        return null;
      case 22:
      case 23:
        return l.lanes = 0, Kr(t, l, e);
      case 24:
        pe(l, Xt, t.memoizedState.cache);
    }
    return re(t, l, e);
  }
  function to(t, l, e) {
    if (t !== null)
      if (t.memoizedProps !== l.pendingProps)
        Kt = !0;
      else {
        if (!Ci(t, e) && (l.flags & 128) === 0)
          return Kt = !1, Ih(
            t,
            l,
            e
          );
        Kt = (t.flags & 131072) !== 0;
      }
    else
      Kt = !1, dt && (l.flags & 1048576) !== 0 && Os(l, gn, l.index);
    switch (l.lanes = 0, l.tag) {
      case 16:
        t: {
          t = l.pendingProps;
          var a = l.elementType, u = a._init;
          if (a = u(a._payload), l.type = a, typeof a == "function")
            Lc(a) ? (t = aa(a, t), l.tag = 1, l = $r(
              null,
              l,
              a,
              t,
              e
            )) : (l.tag = 0, l = Mi(
              null,
              l,
              a,
              t,
              e
            ));
          else {
            if (a != null) {
              if (u = a.$$typeof, u === Lt) {
                l.tag = 11, l = Lr(
                  null,
                  l,
                  a,
                  t,
                  e
                );
                break t;
              } else if (u === Z) {
                l.tag = 14, l = Vr(
                  null,
                  l,
                  a,
                  t,
                  e
                );
                break t;
              }
            }
            throw l = Il(a) || a, Error(r(306, l, ""));
          }
        }
        return l;
      case 0:
        return Mi(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 1:
        return a = l.type, u = aa(
          a,
          l.pendingProps
        ), $r(
          t,
          l,
          a,
          u,
          e
        );
      case 3:
        t: {
          if (pt(
            l,
            l.stateNode.containerInfo
          ), t === null) throw Error(r(387));
          a = l.pendingProps;
          var n = l.memoizedState;
          u = n.element, ai(t, l), xu(l, a, null, e);
          var c = l.memoizedState;
          if (a = c.cache, pe(l, Xt, a), a !== n.cache && Fc(
            l,
            [Xt],
            e,
            !0
          ), bu(), a = c.element, n.isDehydrated)
            if (n = {
              element: a,
              isDehydrated: !1,
              cache: c.cache
            }, l.updateQueue.baseState = n, l.memoizedState = n, l.flags & 256) {
              l = Wr(
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
              ), du(u), l = Wr(
                t,
                l,
                a,
                e
              );
              break t;
            } else
              for (t = l.stateNode.containerInfo, t.nodeType === 9 ? t = t.body : t = t.nodeName === "HTML" ? t.ownerDocument.body : t, _t = Ql(t.firstChild), nl = l, dt = !0, Fe = null, kl = !0, e = _r(
                l,
                null,
                a,
                e
              ), l.child = e; e; )
                e.flags = e.flags & -3 | 4096, e = e.sibling;
          else {
            if (ou(), a === u) {
              l = re(
                t,
                l,
                e
              );
              break t;
            }
            Wt(
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
        return Bn(t, l), t === null ? (e = nd(
          l.type,
          null,
          l.pendingProps,
          null
        )) ? l.memoizedState = e : dt || (e = l.type, t = l.pendingProps, a = Pn(
          $.current
        ).createElement(e), a[It] = l, a[sl] = t, Pt(a, e, t), wt(a), l.stateNode = a) : l.memoizedState = nd(
          l.type,
          t.memoizedProps,
          l.pendingProps,
          t.memoizedState
        ), null;
      case 27:
        return Bl(l), t === null && dt && (a = l.stateNode = ed(
          l.type,
          l.pendingProps,
          $.current
        ), nl = l, kl = !0, u = _t, qe(l.type) ? (vf = u, _t = Ql(
          a.firstChild
        )) : _t = u), Wt(
          t,
          l,
          l.pendingProps.children,
          e
        ), Bn(t, l), t === null && (l.flags |= 4194304), l.child;
      case 5:
        return t === null && dt && ((u = a = _t) && (a = M0(
          a,
          l.type,
          l.pendingProps,
          kl
        ), a !== null ? (l.stateNode = a, nl = l, _t = Ql(
          a.firstChild
        ), kl = !1, u = !0) : u = !1), u || Pe(l)), Bl(l), u = l.type, n = l.pendingProps, c = t !== null ? t.memoizedProps : null, a = n.children, of(u, n) ? a = null : c !== null && of(u, c) && (l.flags |= 32), l.memoizedState !== null && (u = si(
          t,
          l,
          wh,
          null,
          null,
          e
        ), Zu._currentValue = u), Bn(t, l), Wt(t, l, a, e), l.child;
      case 6:
        return t === null && dt && ((t = e = _t) && (e = N0(
          e,
          l.pendingProps,
          kl
        ), e !== null ? (l.stateNode = e, nl = l, _t = null, t = !0) : t = !1), t || Pe(l)), null;
      case 13:
        return Fr(t, l, e);
      case 4:
        return pt(
          l,
          l.stateNode.containerInfo
        ), a = l.pendingProps, t === null ? l.child = Ba(
          l,
          null,
          a,
          e
        ) : Wt(
          t,
          l,
          a,
          e
        ), l.child;
      case 11:
        return Lr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 7:
        return Wt(
          t,
          l,
          l.pendingProps,
          e
        ), l.child;
      case 8:
        return Wt(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 12:
        return Wt(
          t,
          l,
          l.pendingProps.children,
          e
        ), l.child;
      case 10:
        return a = l.pendingProps, pe(l, l.type, a.value), Wt(
          t,
          l,
          a.children,
          e
        ), l.child;
      case 9:
        return u = l.type._context, a = l.pendingProps.children, ta(l), u = tl(u), a = a(u), l.flags |= 1, Wt(t, l, a, e), l.child;
      case 14:
        return Vr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 15:
        return wr(
          t,
          l,
          l.type,
          l.pendingProps,
          e
        );
      case 19:
        return Ir(t, l, e);
      case 31:
        return a = l.pendingProps, e = l.mode, a = {
          mode: a.mode,
          children: a.children
        }, t === null ? (e = Yn(
          a,
          e
        ), e.ref = l.ref, l.child = e, e.return = l, l = e) : (e = ae(t.child, a), e.ref = l.ref, l.child = e, e.return = l, l = e), l;
      case 22:
        return Kr(t, l, e);
      case 24:
        return ta(l), a = tl(Xt), t === null ? (u = ti(), u === null && (u = Tt, n = Pc(), u.pooledCache = n, n.refCount++, n !== null && (u.pooledCacheLanes |= e), u = n), l.memoizedState = {
          parent: a,
          cache: u
        }, ei(l), pe(l, Xt, u)) : ((t.lanes & e) !== 0 && (ai(t, l), xu(l, null, null, e), bu()), u = t.memoizedState, n = l.memoizedState, u.parent !== a ? (u = { parent: a, cache: a }, l.memoizedState = u, l.lanes === 0 && (l.memoizedState = l.updateQueue.baseState = u), pe(l, Xt, a)) : (a = n.cache, pe(l, Xt, a), a !== u.cache && Fc(
          l,
          [Xt],
          e,
          !0
        ))), Wt(
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
  function oe(t) {
    t.flags |= 4;
  }
  function lo(t, l) {
    if (l.type !== "stylesheet" || (l.state.loading & 4) !== 0)
      t.flags &= -16777217;
    else if (t.flags |= 16777216, !rd(l)) {
      if (l = Cl.current, l !== null && ((ct & 4194048) === ct ? $l !== null : (ct & 62914560) !== ct && (ct & 536870912) === 0 || l !== $l))
        throw gu = li, Ys;
      t.flags |= 8192;
    }
  }
  function Gn(t, l) {
    l !== null && (t.flags |= 4), t.flags & 16384 && (l = t.tag !== 22 ? jt() : 536870912, t.lanes |= l, Qa |= l);
  }
  function Mu(t, l) {
    if (!dt)
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
    switch (Jc(l), l.tag) {
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
        return e = l.stateNode, a = null, t !== null && (a = t.memoizedState.cache), l.memoizedState.cache !== a && (l.flags |= 2048), ie(Xt), El(), e.pendingContext && (e.context = e.pendingContext, e.pendingContext = null), (t === null || t.child === null) && (ru(l) ? oe(l) : t === null || t.memoizedState.isDehydrated && (l.flags & 256) === 0 || (l.flags |= 1024, Us())), Nt(l), null;
      case 26:
        return e = l.memoizedState, t === null ? (oe(l), e !== null ? (Nt(l), lo(l, e)) : (Nt(l), l.flags &= -16777217)) : e ? e !== t.memoizedState ? (oe(l), Nt(l), lo(l, e)) : (Nt(l), l.flags &= -16777217) : (t.memoizedProps !== a && oe(l), Nt(l), l.flags &= -16777217), null;
      case 27:
        Ze(l), e = $.current;
        var u = l.type;
        if (t !== null && l.stateNode != null)
          t.memoizedProps !== a && oe(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(r(166));
            return Nt(l), null;
          }
          t = w.current, ru(l) ? Rs(l) : (t = ed(u, a, e), l.stateNode = t, oe(l));
        }
        return Nt(l), null;
      case 5:
        if (Ze(l), e = l.type, t !== null && l.stateNode != null)
          t.memoizedProps !== a && oe(l);
        else {
          if (!a) {
            if (l.stateNode === null)
              throw Error(r(166));
            return Nt(l), null;
          }
          if (t = w.current, ru(l))
            Rs(l);
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
            t[It] = l, t[sl] = a;
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
            t: switch (Pt(t, e, a), e) {
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
            t && oe(l);
          }
        }
        return Nt(l), l.flags &= -16777217, null;
      case 6:
        if (t && l.stateNode != null)
          t.memoizedProps !== a && oe(l);
        else {
          if (typeof a != "string" && l.stateNode === null)
            throw Error(r(166));
          if (t = $.current, ru(l)) {
            if (t = l.stateNode, e = l.memoizedProps, a = null, u = nl, u !== null)
              switch (u.tag) {
                case 27:
                case 5:
                  a = u.memoizedProps;
              }
            t[It] = l, t = !!(t.nodeValue === e || a !== null && a.suppressHydrationWarning === !0 || $o(t.nodeValue, e)), t || Pe(l);
          } else
            t = Pn(t).createTextNode(
              a
            ), t[It] = l, l.stateNode = t;
        }
        return Nt(l), null;
      case 13:
        if (a = l.memoizedState, t === null || t.memoizedState !== null && t.memoizedState.dehydrated !== null) {
          if (u = ru(l), a !== null && a.dehydrated !== null) {
            if (t === null) {
              if (!u) throw Error(r(318));
              if (u = l.memoizedState, u = u !== null ? u.dehydrated : null, !u) throw Error(r(317));
              u[It] = l;
            } else
              ou(), (l.flags & 128) === 0 && (l.memoizedState = null), l.flags |= 4;
            Nt(l), u = !1;
          } else
            u = Us(), t !== null && t.memoizedState !== null && (t.memoizedState.hydrationErrors = u), u = !0;
          if (!u)
            return l.flags & 256 ? (se(l), l) : (se(l), null);
        }
        if (se(l), (l.flags & 128) !== 0)
          return l.lanes = e, l;
        if (e = a !== null, t = t !== null && t.memoizedState !== null, e) {
          a = l.child, u = null, a.alternate !== null && a.alternate.memoizedState !== null && a.alternate.memoizedState.cachePool !== null && (u = a.alternate.memoizedState.cachePool.pool);
          var n = null;
          a.memoizedState !== null && a.memoizedState.cachePool !== null && (n = a.memoizedState.cachePool.pool), n !== u && (a.flags |= 2048);
        }
        return e !== t && e && (l.child.flags |= 8192), Gn(l, l.updateQueue), Nt(l), null;
      case 4:
        return El(), t === null && nf(l.stateNode.containerInfo), Nt(l), null;
      case 10:
        return ie(l.type), Nt(l), null;
      case 19:
        if (Y(Qt), u = l.memoizedState, u === null) return Nt(l), null;
        if (a = (l.flags & 128) !== 0, n = u.rendering, n === null)
          if (a) Mu(u, !1);
          else {
            if (Ut !== 0 || t !== null && (t.flags & 128) !== 0)
              for (t = l.child; t !== null; ) {
                if (n = Cn(t), n !== null) {
                  for (l.flags |= 128, Mu(u, !1), t = n.updateQueue, l.updateQueue = t, Gn(l, t), l.subtreeFlags = 0, t = e, e = l.child; e !== null; )
                    Ds(e, t), e = e.sibling;
                  return H(
                    Qt,
                    Qt.current & 1 | 2
                  ), l.child;
                }
                t = t.sibling;
              }
            u.tail !== null && al() > Zn && (l.flags |= 128, a = !0, Mu(u, !1), l.lanes = 4194304);
          }
        else {
          if (!a)
            if (t = Cn(n), t !== null) {
              if (l.flags |= 128, a = !0, t = t.updateQueue, l.updateQueue = t, Gn(l, t), Mu(u, !0), u.tail === null && u.tailMode === "hidden" && !n.alternate && !dt)
                return Nt(l), null;
            } else
              2 * al() - u.renderingStartTime > Zn && e !== 536870912 && (l.flags |= 128, a = !0, Mu(u, !1), l.lanes = 4194304);
          u.isBackwards ? (n.sibling = l.child, l.child = n) : (t = u.last, t !== null ? t.sibling = n : l.child = n, u.last = n);
        }
        return u.tail !== null ? (l = u.tail, u.rendering = l, u.tail = l.sibling, u.renderingStartTime = al(), l.sibling = null, t = Qt.current, H(Qt, a ? t & 1 | 2 : t & 1), l) : (Nt(l), null);
      case 22:
      case 23:
        return se(l), ii(), a = l.memoizedState !== null, t !== null ? t.memoizedState !== null !== a && (l.flags |= 8192) : a && (l.flags |= 8192), a ? (e & 536870912) !== 0 && (l.flags & 128) === 0 && (Nt(l), l.subtreeFlags & 6 && (l.flags |= 8192)) : Nt(l), e = l.updateQueue, e !== null && Gn(l, e.retryQueue), e = null, t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), a = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (a = l.memoizedState.cachePool.pool), a !== e && (l.flags |= 2048), t !== null && Y(la), null;
      case 24:
        return e = null, t !== null && (e = t.memoizedState.cache), l.memoizedState.cache !== e && (l.flags |= 2048), ie(Xt), Nt(l), null;
      case 25:
        return null;
      case 30:
        return null;
    }
    throw Error(r(156, l.tag));
  }
  function l0(t, l) {
    switch (Jc(l), l.tag) {
      case 1:
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 3:
        return ie(Xt), El(), t = l.flags, (t & 65536) !== 0 && (t & 128) === 0 ? (l.flags = t & -65537 | 128, l) : null;
      case 26:
      case 27:
      case 5:
        return Ze(l), null;
      case 13:
        if (se(l), t = l.memoizedState, t !== null && t.dehydrated !== null) {
          if (l.alternate === null)
            throw Error(r(340));
          ou();
        }
        return t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 19:
        return Y(Qt), null;
      case 4:
        return El(), null;
      case 10:
        return ie(l.type), null;
      case 22:
      case 23:
        return se(l), ii(), t !== null && Y(la), t = l.flags, t & 65536 ? (l.flags = t & -65537 | 128, l) : null;
      case 24:
        return ie(Xt), null;
      case 25:
        return null;
      default:
        return null;
    }
  }
  function eo(t, l) {
    switch (Jc(l), l.tag) {
      case 3:
        ie(Xt), El();
        break;
      case 26:
      case 27:
      case 5:
        Ze(l);
        break;
      case 4:
        El();
        break;
      case 13:
        se(l);
        break;
      case 19:
        Y(Qt);
        break;
      case 10:
        ie(l.type);
        break;
      case 22:
      case 23:
        se(l), ii(), t !== null && Y(la);
        break;
      case 24:
        ie(Xt);
    }
  }
  function Nu(t, l) {
    try {
      var e = l.updateQueue, a = e !== null ? e.lastEffect : null;
      if (a !== null) {
        var u = a.next;
        e = u;
        do {
          if ((e.tag & t) === t) {
            a = void 0;
            var n = e.create, c = e.inst;
            a = n(), c.destroy = a;
          }
          e = e.next;
        } while (e !== u);
      }
    } catch (f) {
      St(l, l.return, f);
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
            var c = a.inst, f = c.destroy;
            if (f !== void 0) {
              c.destroy = void 0, u = l;
              var s = e, m = f;
              try {
                m();
              } catch (j) {
                St(
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
      St(l, l.return, j);
    }
  }
  function ao(t) {
    var l = t.updateQueue;
    if (l !== null) {
      var e = t.stateNode;
      try {
        Vs(l, e);
      } catch (a) {
        St(t, t.return, a);
      }
    }
  }
  function uo(t, l, e) {
    e.props = aa(
      t.type,
      t.memoizedProps
    ), e.state = t.memoizedState;
    try {
      e.componentWillUnmount();
    } catch (a) {
      St(t, l, a);
    }
  }
  function Du(t, l) {
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
      St(t, l, u);
    }
  }
  function Wl(t, l) {
    var e = t.ref, a = t.refCleanup;
    if (e !== null)
      if (typeof a == "function")
        try {
          a();
        } catch (u) {
          St(t, l, u);
        } finally {
          t.refCleanup = null, t = t.alternate, t != null && (t.refCleanup = null);
        }
      else if (typeof e == "function")
        try {
          e(null);
        } catch (u) {
          St(t, l, u);
        }
      else e.current = null;
  }
  function no(t) {
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
      St(t, t.return, u);
    }
  }
  function Hi(t, l, e) {
    try {
      var a = t.stateNode;
      j0(a, t.type, e, l), a[sl] = l;
    } catch (u) {
      St(t, t.return, u);
    }
  }
  function co(t) {
    return t.tag === 5 || t.tag === 3 || t.tag === 26 || t.tag === 27 && qe(t.type) || t.tag === 4;
  }
  function qi(t) {
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
  function Bi(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? (e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e).insertBefore(t, l) : (l = e.nodeType === 9 ? e.body : e.nodeName === "HTML" ? e.ownerDocument.body : e, l.appendChild(t), e = e._reactRootContainer, e != null || l.onclick !== null || (l.onclick = Fn));
    else if (a !== 4 && (a === 27 && qe(t.type) && (e = t.stateNode, l = null), t = t.child, t !== null))
      for (Bi(t, l, e), t = t.sibling; t !== null; )
        Bi(t, l, e), t = t.sibling;
  }
  function Xn(t, l, e) {
    var a = t.tag;
    if (a === 5 || a === 6)
      t = t.stateNode, l ? e.insertBefore(t, l) : e.appendChild(t);
    else if (a !== 4 && (a === 27 && qe(t.type) && (e = t.stateNode), t = t.child, t !== null))
      for (Xn(t, l, e), t = t.sibling; t !== null; )
        Xn(t, l, e), t = t.sibling;
  }
  function io(t) {
    var l = t.stateNode, e = t.memoizedProps;
    try {
      for (var a = t.type, u = l.attributes; u.length; )
        l.removeAttributeNode(u[0]);
      Pt(l, a, e), l[It] = t, l[sl] = e;
    } catch (n) {
      St(t, t.return, n);
    }
  }
  var de = !1, qt = !1, Yi = !1, fo = typeof WeakSet == "function" ? WeakSet : Set, Jt = null;
  function e0(t, l) {
    if (t = t.containerInfo, sf = uc, t = bs(t), qc(t)) {
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
            var c = 0, f = -1, s = -1, m = 0, j = 0, A = t, g = null;
            l: for (; ; ) {
              for (var S; A !== e || u !== 0 && A.nodeType !== 3 || (f = c + u), A !== n || a !== 0 && A.nodeType !== 3 || (s = c + a), A.nodeType === 3 && (c += A.nodeValue.length), (S = A.firstChild) !== null; )
                g = A, A = S;
              for (; ; ) {
                if (A === t) break l;
                if (g === e && ++m === u && (f = c), g === n && ++j === a && (s = c), (S = A.nextSibling) !== null) break;
                A = g, g = A.parentNode;
              }
              A = S;
            }
            e = f === -1 || s === -1 ? null : { start: f, end: s };
          } else e = null;
        }
      e = e || { start: 0, end: 0 };
    } else e = null;
    for (rf = { focusedElem: t, selectionRange: e }, uc = !1, Jt = l; Jt !== null; )
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
                  var k = aa(
                    e.type,
                    u,
                    e.elementType === e.type
                  );
                  t = a.getSnapshotBeforeUpdate(
                    k,
                    n
                  ), a.__reactInternalSnapshotBeforeUpdate = t;
                } catch (K) {
                  St(
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
                  hf(t);
                else if (e === 1)
                  switch (t.nodeName) {
                    case "HEAD":
                    case "HTML":
                    case "BODY":
                      hf(t);
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
            t.return = l.return, Jt = t;
            break;
          }
          Jt = l.return;
        }
  }
  function so(t, l, e) {
    var a = e.flags;
    switch (e.tag) {
      case 0:
      case 11:
      case 15:
        De(t, e), a & 4 && Nu(5, e);
        break;
      case 1:
        if (De(t, e), a & 4)
          if (t = e.stateNode, l === null)
            try {
              t.componentDidMount();
            } catch (c) {
              St(e, e.return, c);
            }
          else {
            var u = aa(
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
            } catch (c) {
              St(
                e,
                e.return,
                c
              );
            }
          }
        a & 64 && ao(e), a & 512 && Du(e, e.return);
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
            Vs(t, l);
          } catch (c) {
            St(e, e.return, c);
          }
        }
        break;
      case 27:
        l === null && a & 4 && io(e);
      case 26:
      case 5:
        De(t, e), l === null && a & 4 && no(e), a & 512 && Du(e, e.return);
        break;
      case 12:
        De(t, e);
        break;
      case 13:
        De(t, e), a & 4 && ho(t, e), a & 64 && (t = e.memoizedState, t !== null && (t = t.dehydrated, t !== null && (e = o0.bind(
          null,
          e
        ), D0(t, e))));
        break;
      case 22:
        if (a = e.memoizedState !== null || de, !a) {
          l = l !== null && l.memoizedState !== null || qt, u = de;
          var n = qt;
          de = a, (qt = l) && !n ? Oe(
            t,
            e,
            (e.subtreeFlags & 8772) !== 0
          ) : De(t, e), de = u, qt = n;
        }
        break;
      case 30:
        break;
      default:
        De(t, e);
    }
  }
  function ro(t) {
    var l = t.alternate;
    l !== null && (t.alternate = null, ro(l)), t.child = null, t.deletions = null, t.sibling = null, t.tag === 5 && (l = t.stateNode, l !== null && gc(l)), t.stateNode = null, t.return = null, t.dependencies = null, t.memoizedProps = null, t.memoizedState = null, t.pendingProps = null, t.stateNode = null, t.updateQueue = null;
  }
  var Mt = null, dl = !1;
  function he(t, l, e) {
    for (e = e.child; e !== null; )
      oo(t, l, e), e = e.sibling;
  }
  function oo(t, l, e) {
    if (nt && typeof nt.onCommitFiberUnmount == "function")
      try {
        nt.onCommitFiberUnmount(zl, e);
      } catch {
      }
    switch (e.tag) {
      case 26:
        qt || Wl(e, l), he(
          t,
          l,
          e
        ), e.memoizedState ? e.memoizedState.count-- : e.stateNode && (e = e.stateNode, e.parentNode.removeChild(e));
        break;
      case 27:
        qt || Wl(e, l);
        var a = Mt, u = dl;
        qe(e.type) && (Mt = e.stateNode, dl = !1), he(
          t,
          l,
          e
        ), Yu(e.stateNode), Mt = a, dl = u;
        break;
      case 5:
        qt || Wl(e, l);
      case 6:
        if (a = Mt, u = dl, Mt = null, he(
          t,
          l,
          e
        ), Mt = a, dl = u, Mt !== null)
          if (dl)
            try {
              (Mt.nodeType === 9 ? Mt.body : Mt.nodeName === "HTML" ? Mt.ownerDocument.body : Mt).removeChild(e.stateNode);
            } catch (n) {
              St(
                e,
                l,
                n
              );
            }
          else
            try {
              Mt.removeChild(e.stateNode);
            } catch (n) {
              St(
                e,
                l,
                n
              );
            }
        break;
      case 18:
        Mt !== null && (dl ? (t = Mt, td(
          t.nodeType === 9 ? t.body : t.nodeName === "HTML" ? t.ownerDocument.body : t,
          e.stateNode
        ), Ku(t)) : td(Mt, e.stateNode));
        break;
      case 4:
        a = Mt, u = dl, Mt = e.stateNode.containerInfo, dl = !0, he(
          t,
          l,
          e
        ), Mt = a, dl = u;
        break;
      case 0:
      case 11:
      case 14:
      case 15:
        qt || Ne(2, e, l), qt || Ne(4, e, l), he(
          t,
          l,
          e
        );
        break;
      case 1:
        qt || (Wl(e, l), a = e.stateNode, typeof a.componentWillUnmount == "function" && uo(
          e,
          l,
          a
        )), he(
          t,
          l,
          e
        );
        break;
      case 21:
        he(
          t,
          l,
          e
        );
        break;
      case 22:
        qt = (a = qt) || e.memoizedState !== null, he(
          t,
          l,
          e
        ), qt = a;
        break;
      default:
        he(
          t,
          l,
          e
        );
    }
  }
  function ho(t, l) {
    if (l.memoizedState === null && (t = l.alternate, t !== null && (t = t.memoizedState, t !== null && (t = t.dehydrated, t !== null))))
      try {
        Ku(t);
      } catch (e) {
        St(l, l.return, e);
      }
  }
  function a0(t) {
    switch (t.tag) {
      case 13:
      case 19:
        var l = t.stateNode;
        return l === null && (l = t.stateNode = new fo()), l;
      case 22:
        return t = t.stateNode, l = t._retryCache, l === null && (l = t._retryCache = new fo()), l;
      default:
        throw Error(r(435, t.tag));
    }
  }
  function Gi(t, l) {
    var e = a0(t);
    l.forEach(function(a) {
      var u = d0.bind(null, t, a);
      e.has(a) || (e.add(a), a.then(u, u));
    });
  }
  function bl(t, l) {
    var e = l.deletions;
    if (e !== null)
      for (var a = 0; a < e.length; a++) {
        var u = e[a], n = t, c = l, f = c;
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
        if (Mt === null) throw Error(r(160));
        oo(n, c, u), Mt = null, dl = !1, n = u.alternate, n !== null && (n.return = null), u.return = null;
      }
    if (l.subtreeFlags & 13878)
      for (l = l.child; l !== null; )
        yo(l, t), l = l.sibling;
  }
  var Xl = null;
  function yo(t, l) {
    var e = t.alternate, a = t.flags;
    switch (t.tag) {
      case 0:
      case 11:
      case 14:
      case 15:
        bl(l, t), xl(t), a & 4 && (Ne(3, t, t.return), Nu(3, t), Ne(5, t, t.return));
        break;
      case 1:
        bl(l, t), xl(t), a & 512 && (qt || e === null || Wl(e, e.return)), a & 64 && de && (t = t.updateQueue, t !== null && (a = t.callbacks, a !== null && (e = t.shared.hiddenCallbacks, t.shared.hiddenCallbacks = e === null ? a : e.concat(a))));
        break;
      case 26:
        var u = Xl;
        if (bl(l, t), xl(t), a & 512 && (qt || e === null || Wl(e, e.return)), a & 4) {
          var n = e !== null ? e.memoizedState : null;
          if (a = t.memoizedState, e === null)
            if (a === null)
              if (t.stateNode === null) {
                t: {
                  a = t.type, e = t.memoizedProps, u = u.ownerDocument || u;
                  l: switch (a) {
                    case "title":
                      n = u.getElementsByTagName("title")[0], (!n || n[Ia] || n[It] || n.namespaceURI === "http://www.w3.org/2000/svg" || n.hasAttribute("itemprop")) && (n = u.createElement(a), u.head.insertBefore(
                        n,
                        u.querySelector("head > title")
                      )), Pt(n, a, e), n[It] = t, wt(n), a = n;
                      break t;
                    case "link":
                      var c = fd(
                        "link",
                        "href",
                        u
                      ).get(a + (e.href || ""));
                      if (c) {
                        for (var f = 0; f < c.length; f++)
                          if (n = c[f], n.getAttribute("href") === (e.href == null || e.href === "" ? null : e.href) && n.getAttribute("rel") === (e.rel == null ? null : e.rel) && n.getAttribute("title") === (e.title == null ? null : e.title) && n.getAttribute("crossorigin") === (e.crossOrigin == null ? null : e.crossOrigin)) {
                            c.splice(f, 1);
                            break l;
                          }
                      }
                      n = u.createElement(a), Pt(n, a, e), u.head.appendChild(n);
                      break;
                    case "meta":
                      if (c = fd(
                        "meta",
                        "content",
                        u
                      ).get(a + (e.content || ""))) {
                        for (f = 0; f < c.length; f++)
                          if (n = c[f], n.getAttribute("content") === (e.content == null ? null : "" + e.content) && n.getAttribute("name") === (e.name == null ? null : e.name) && n.getAttribute("property") === (e.property == null ? null : e.property) && n.getAttribute("http-equiv") === (e.httpEquiv == null ? null : e.httpEquiv) && n.getAttribute("charset") === (e.charSet == null ? null : e.charSet)) {
                            c.splice(f, 1);
                            break l;
                          }
                      }
                      n = u.createElement(a), Pt(n, a, e), u.head.appendChild(n);
                      break;
                    default:
                      throw Error(r(468, a));
                  }
                  n[It] = t, wt(n), a = n;
                }
                t.stateNode = a;
              } else
                sd(
                  u,
                  t.type,
                  t.stateNode
                );
            else
              t.stateNode = id(
                u,
                a,
                t.memoizedProps
              );
          else
            n !== a ? (n === null ? e.stateNode !== null && (e = e.stateNode, e.parentNode.removeChild(e)) : n.count--, a === null ? sd(
              u,
              t.type,
              t.stateNode
            ) : id(
              u,
              a,
              t.memoizedProps
            )) : a === null && t.stateNode !== null && Hi(
              t,
              t.memoizedProps,
              e.memoizedProps
            );
        }
        break;
      case 27:
        bl(l, t), xl(t), a & 512 && (qt || e === null || Wl(e, e.return)), e !== null && a & 4 && Hi(
          t,
          t.memoizedProps,
          e.memoizedProps
        );
        break;
      case 5:
        if (bl(l, t), xl(t), a & 512 && (qt || e === null || Wl(e, e.return)), t.flags & 32) {
          u = t.stateNode;
          try {
            ba(u, "");
          } catch (S) {
            St(t, t.return, S);
          }
        }
        a & 4 && t.stateNode != null && (u = t.memoizedProps, Hi(
          t,
          u,
          e !== null ? e.memoizedProps : u
        )), a & 1024 && (Yi = !0);
        break;
      case 6:
        if (bl(l, t), xl(t), a & 4) {
          if (t.stateNode === null)
            throw Error(r(162));
          a = t.memoizedProps, e = t.stateNode;
          try {
            e.nodeValue = a;
          } catch (S) {
            St(t, t.return, S);
          }
        }
        break;
      case 3:
        if (lc = null, u = Xl, Xl = In(l.containerInfo), bl(l, t), Xl = u, xl(t), a & 4 && e !== null && e.memoizedState.isDehydrated)
          try {
            Ku(l.containerInfo);
          } catch (S) {
            St(t, t.return, S);
          }
        Yi && (Yi = !1, vo(t));
        break;
      case 4:
        a = Xl, Xl = In(
          t.stateNode.containerInfo
        ), bl(l, t), xl(t), Xl = a;
        break;
      case 12:
        bl(l, t), xl(t);
        break;
      case 13:
        bl(l, t), xl(t), t.child.flags & 8192 && t.memoizedState !== null != (e !== null && e.memoizedState !== null) && (wi = al()), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Gi(t, a)));
        break;
      case 22:
        u = t.memoizedState !== null;
        var s = e !== null && e.memoizedState !== null, m = de, j = qt;
        if (de = m || u, qt = j || s, bl(l, t), qt = j, de = m, xl(t), a & 8192)
          t: for (l = t.stateNode, l._visibility = u ? l._visibility & -2 : l._visibility | 1, u && (e === null || s || de || qt || ua(t)), e = null, l = t; ; ) {
            if (l.tag === 5 || l.tag === 26) {
              if (e === null) {
                s = e = l;
                try {
                  if (n = s.stateNode, u)
                    c = n.style, typeof c.setProperty == "function" ? c.setProperty("display", "none", "important") : c.display = "none";
                  else {
                    f = s.stateNode;
                    var A = s.memoizedProps.style, g = A != null && A.hasOwnProperty("display") ? A.display : null;
                    f.style.display = g == null || typeof g == "boolean" ? "" : ("" + g).trim();
                  }
                } catch (S) {
                  St(s, s.return, S);
                }
              }
            } else if (l.tag === 6) {
              if (e === null) {
                s = l;
                try {
                  s.stateNode.nodeValue = u ? "" : s.memoizedProps;
                } catch (S) {
                  St(s, s.return, S);
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
        a & 4 && (a = t.updateQueue, a !== null && (e = a.retryQueue, e !== null && (a.retryQueue = null, Gi(t, e))));
        break;
      case 19:
        bl(l, t), xl(t), a & 4 && (a = t.updateQueue, a !== null && (t.updateQueue = null, Gi(t, a)));
        break;
      case 30:
        break;
      case 21:
        break;
      default:
        bl(l, t), xl(t);
    }
  }
  function xl(t) {
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
            var u = e.stateNode, n = qi(t);
            Xn(t, n, u);
            break;
          case 5:
            var c = e.stateNode;
            e.flags & 32 && (ba(c, ""), e.flags &= -33);
            var f = qi(t);
            Xn(t, f, c);
            break;
          case 3:
          case 4:
            var s = e.stateNode.containerInfo, m = qi(t);
            Bi(
              t,
              m,
              s
            );
            break;
          default:
            throw Error(r(161));
        }
      } catch (j) {
        St(t, t.return, j);
      }
      t.flags &= -3;
    }
    l & 4096 && (t.flags &= -4097);
  }
  function vo(t) {
    if (t.subtreeFlags & 1024)
      for (t = t.child; t !== null; ) {
        var l = t;
        vo(l), l.tag === 5 && l.flags & 1024 && l.stateNode.reset(), t = t.sibling;
      }
  }
  function De(t, l) {
    if (l.subtreeFlags & 8772)
      for (l = l.child; l !== null; )
        so(t, l.alternate, l), l = l.sibling;
  }
  function ua(t) {
    for (t = t.child; t !== null; ) {
      var l = t;
      switch (l.tag) {
        case 0:
        case 11:
        case 14:
        case 15:
          Ne(4, l, l.return), ua(l);
          break;
        case 1:
          Wl(l, l.return);
          var e = l.stateNode;
          typeof e.componentWillUnmount == "function" && uo(
            l,
            l.return,
            e
          ), ua(l);
          break;
        case 27:
          Yu(l.stateNode);
        case 26:
        case 5:
          Wl(l, l.return), ua(l);
          break;
        case 22:
          l.memoizedState === null && ua(l);
          break;
        case 30:
          ua(l);
          break;
        default:
          ua(l);
      }
      t = t.sibling;
    }
  }
  function Oe(t, l, e) {
    for (e = e && (l.subtreeFlags & 8772) !== 0, l = l.child; l !== null; ) {
      var a = l.alternate, u = t, n = l, c = n.flags;
      switch (n.tag) {
        case 0:
        case 11:
        case 15:
          Oe(
            u,
            n,
            e
          ), Nu(4, n);
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
              St(a, a.return, m);
            }
          if (a = n, u = a.updateQueue, u !== null) {
            var f = a.stateNode;
            try {
              var s = u.shared.hiddenCallbacks;
              if (s !== null)
                for (u.shared.hiddenCallbacks = null, u = 0; u < s.length; u++)
                  Ls(s[u], f);
            } catch (m) {
              St(a, a.return, m);
            }
          }
          e && c & 64 && ao(n), Du(n, n.return);
          break;
        case 27:
          io(n);
        case 26:
        case 5:
          Oe(
            u,
            n,
            e
          ), e && a === null && c & 4 && no(n), Du(n, n.return);
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
          ), e && c & 4 && ho(u, n);
          break;
        case 22:
          n.memoizedState === null && Oe(
            u,
            n,
            e
          ), Du(n, n.return);
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
  function Xi(t, l) {
    var e = null;
    t !== null && t.memoizedState !== null && t.memoizedState.cachePool !== null && (e = t.memoizedState.cachePool.pool), t = null, l.memoizedState !== null && l.memoizedState.cachePool !== null && (t = l.memoizedState.cachePool.pool), t !== e && (t != null && t.refCount++, e != null && yu(e));
  }
  function Qi(t, l) {
    t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && yu(t));
  }
  function Fl(t, l, e, a) {
    if (l.subtreeFlags & 10256)
      for (l = l.child; l !== null; )
        mo(
          t,
          l,
          e,
          a
        ), l = l.sibling;
  }
  function mo(t, l, e, a) {
    var u = l.flags;
    switch (l.tag) {
      case 0:
      case 11:
      case 15:
        Fl(
          t,
          l,
          e,
          a
        ), u & 2048 && Nu(9, l);
        break;
      case 1:
        Fl(
          t,
          l,
          e,
          a
        );
        break;
      case 3:
        Fl(
          t,
          l,
          e,
          a
        ), u & 2048 && (t = null, l.alternate !== null && (t = l.alternate.memoizedState.cache), l = l.memoizedState.cache, l !== t && (l.refCount++, t != null && yu(t)));
        break;
      case 12:
        if (u & 2048) {
          Fl(
            t,
            l,
            e,
            a
          ), t = l.stateNode;
          try {
            var n = l.memoizedProps, c = n.id, f = n.onPostCommit;
            typeof f == "function" && f(
              c,
              l.alternate === null ? "mount" : "update",
              t.passiveEffectDuration,
              -0
            );
          } catch (s) {
            St(l, l.return, s);
          }
        } else
          Fl(
            t,
            l,
            e,
            a
          );
        break;
      case 13:
        Fl(
          t,
          l,
          e,
          a
        );
        break;
      case 23:
        break;
      case 22:
        n = l.stateNode, c = l.alternate, l.memoizedState !== null ? n._visibility & 2 ? Fl(
          t,
          l,
          e,
          a
        ) : Ou(t, l) : n._visibility & 2 ? Fl(
          t,
          l,
          e,
          a
        ) : (n._visibility |= 2, Ya(
          t,
          l,
          e,
          a,
          (l.subtreeFlags & 10256) !== 0
        )), u & 2048 && Xi(c, l);
        break;
      case 24:
        Fl(
          t,
          l,
          e,
          a
        ), u & 2048 && Qi(l.alternate, l);
        break;
      default:
        Fl(
          t,
          l,
          e,
          a
        );
    }
  }
  function Ya(t, l, e, a, u) {
    for (u = u && (l.subtreeFlags & 10256) !== 0, l = l.child; l !== null; ) {
      var n = t, c = l, f = e, s = a, m = c.flags;
      switch (c.tag) {
        case 0:
        case 11:
        case 15:
          Ya(
            n,
            c,
            f,
            s,
            u
          ), Nu(8, c);
          break;
        case 23:
          break;
        case 22:
          var j = c.stateNode;
          c.memoizedState !== null ? j._visibility & 2 ? Ya(
            n,
            c,
            f,
            s,
            u
          ) : Ou(
            n,
            c
          ) : (j._visibility |= 2, Ya(
            n,
            c,
            f,
            s,
            u
          )), u && m & 2048 && Xi(
            c.alternate,
            c
          );
          break;
        case 24:
          Ya(
            n,
            c,
            f,
            s,
            u
          ), u && m & 2048 && Qi(c.alternate, c);
          break;
        default:
          Ya(
            n,
            c,
            f,
            s,
            u
          );
      }
      l = l.sibling;
    }
  }
  function Ou(t, l) {
    if (l.subtreeFlags & 10256)
      for (l = l.child; l !== null; ) {
        var e = t, a = l, u = a.flags;
        switch (a.tag) {
          case 22:
            Ou(e, a), u & 2048 && Xi(
              a.alternate,
              a
            );
            break;
          case 24:
            Ou(e, a), u & 2048 && Qi(a.alternate, a);
            break;
          default:
            Ou(e, a);
        }
        l = l.sibling;
      }
  }
  var Ru = 8192;
  function Ga(t) {
    if (t.subtreeFlags & Ru)
      for (t = t.child; t !== null; )
        go(t), t = t.sibling;
  }
  function go(t) {
    switch (t.tag) {
      case 26:
        Ga(t), t.flags & Ru && t.memoizedState !== null && Z0(
          Xl,
          t.memoizedState,
          t.memoizedProps
        );
        break;
      case 5:
        Ga(t);
        break;
      case 3:
      case 4:
        var l = Xl;
        Xl = In(t.stateNode.containerInfo), Ga(t), Xl = l;
        break;
      case 22:
        t.memoizedState === null && (l = t.alternate, l !== null && l.memoizedState !== null ? (l = Ru, Ru = 16777216, Ga(t), Ru = l) : Ga(t));
        break;
      default:
        Ga(t);
    }
  }
  function So(t) {
    var l = t.alternate;
    if (l !== null && (t = l.child, t !== null)) {
      l.child = null;
      do
        l = t.sibling, t.sibling = null, t = l;
      while (t !== null);
    }
  }
  function _u(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          Jt = a, xo(
            a,
            t
          );
        }
      So(t);
    }
    if (t.subtreeFlags & 10256)
      for (t = t.child; t !== null; )
        bo(t), t = t.sibling;
  }
  function bo(t) {
    switch (t.tag) {
      case 0:
      case 11:
      case 15:
        _u(t), t.flags & 2048 && Ne(9, t, t.return);
        break;
      case 3:
        _u(t);
        break;
      case 12:
        _u(t);
        break;
      case 22:
        var l = t.stateNode;
        t.memoizedState !== null && l._visibility & 2 && (t.return === null || t.return.tag !== 13) ? (l._visibility &= -3, Qn(t)) : _u(t);
        break;
      default:
        _u(t);
    }
  }
  function Qn(t) {
    var l = t.deletions;
    if ((t.flags & 16) !== 0) {
      if (l !== null)
        for (var e = 0; e < l.length; e++) {
          var a = l[e];
          Jt = a, xo(
            a,
            t
          );
        }
      So(t);
    }
    for (t = t.child; t !== null; ) {
      switch (l = t, l.tag) {
        case 0:
        case 11:
        case 15:
          Ne(8, l, l.return), Qn(l);
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
          yu(e.memoizedState.cache);
      }
      if (a = e.child, a !== null) a.return = e, Jt = a;
      else
        t: for (e = t; Jt !== null; ) {
          a = Jt;
          var u = a.sibling, n = a.return;
          if (ro(a), a === e) {
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
      var l = tl(Xt), e = l.data.get(t);
      return e === void 0 && (e = t(), l.data.set(t, e)), e;
    }
  }, n0 = typeof WeakMap == "function" ? WeakMap : Map, ht = 0, Tt = null, et = null, ct = 0, yt = 0, pl = null, Re = !1, Xa = !1, Zi = !1, ye = 0, Ut = 0, _e = 0, na = 0, Li = 0, Hl = 0, Qa = 0, Uu = null, hl = null, Vi = !1, wi = 0, Zn = 1 / 0, Ln = null, Ue = null, Ft = 0, Ce = null, Za = null, La = 0, Ki = 0, Ji = null, po = null, Cu = 0, ki = null;
  function jl() {
    if ((ht & 2) !== 0 && ct !== 0)
      return ct & -ct;
    if (p.T !== null) {
      var t = Oa;
      return t !== 0 ? t : lf();
    }
    return Iu();
  }
  function jo() {
    Hl === 0 && (Hl = (ct & 536870912) === 0 || dt ? F() : 536870912);
    var t = Cl.current;
    return t !== null && (t.flags |= 32), Hl;
  }
  function Tl(t, l, e) {
    (t === Tt && (yt === 2 || yt === 9) || t.cancelPendingCommit !== null) && (Va(t, 0), He(
      t,
      ct,
      Hl,
      !1
    )), Ct(t, e), ((ht & 2) === 0 || t !== Tt) && (t === Tt && ((ht & 2) === 0 && (na |= e), Ut === 4 && He(
      t,
      ct,
      Hl,
      !1
    )), Pl(t));
  }
  function To(t, l, e) {
    if ((ht & 6) !== 0) throw Error(r(327));
    var a = !e && (l & 124) === 0 && (l & t.expiredLanes) === 0 || E(t, l), u = a ? f0(t, l) : Fi(t, l, !0), n = a;
    do {
      if (u === 0) {
        Xa && !a && He(t, l, 0, !1);
        break;
      } else {
        if (e = t.current.alternate, n && !c0(e)) {
          u = Fi(t, l, !1), n = !1;
          continue;
        }
        if (u === 2) {
          if (n = l, t.errorRecoveryDisabledLanes & n)
            var c = 0;
          else
            c = t.pendingLanes & -536870913, c = c !== 0 ? c : c & 536870912 ? 536870912 : 0;
          if (c !== 0) {
            l = c;
            t: {
              var f = t;
              u = Uu;
              var s = f.current.memoizedState.isDehydrated;
              if (s && (Va(f, c).flags |= 256), c = Fi(
                f,
                c,
                !1
              ), c !== 2) {
                if (Zi && !s) {
                  f.errorRecoveryDisabledLanes |= n, na |= n, u = 4;
                  break t;
                }
                n = hl, hl = u, n !== null && (hl === null ? hl = n : hl.push.apply(
                  hl,
                  n
                ));
              }
              u = c;
            }
            if (n = !1, u !== 2) continue;
          }
        }
        if (u === 1) {
          Va(t, 0), He(t, l, 0, !0);
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
              throw Error(r(329));
          }
          if ((l & 62914560) === l && (u = wi + 300 - al(), 10 < u)) {
            if (He(
              a,
              l,
              Hl,
              !Re
            ), wl(a, 0, !0) !== 0) break t;
            a.timeoutHandle = Po(
              Ao.bind(
                null,
                a,
                e,
                hl,
                Ln,
                Vi,
                l,
                Hl,
                na,
                Qa,
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
          Ao(
            a,
            e,
            hl,
            Ln,
            Vi,
            l,
            Hl,
            na,
            Qa,
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
    Pl(t);
  }
  function Ao(t, l, e, a, u, n, c, f, s, m, j, A, g, S) {
    if (t.timeoutHandle = -1, A = l.subtreeFlags, (A & 8192 || (A & 16785408) === 16785408) && (Qu = { stylesheets: null, count: 0, unsuspend: Q0 }, go(l), A = L0(), A !== null)) {
      t.cancelPendingCommit = A(
        Ro.bind(
          null,
          t,
          l,
          n,
          e,
          a,
          u,
          c,
          f,
          s,
          j,
          1,
          g,
          S
        )
      ), He(t, n, c, !m);
      return;
    }
    Ro(
      t,
      l,
      n,
      e,
      a,
      u,
      c,
      f,
      s
    );
  }
  function c0(t) {
    for (var l = t; ; ) {
      var e = l.tag;
      if ((e === 0 || e === 11 || e === 15) && l.flags & 16384 && (e = l.updateQueue, e !== null && (e = e.stores, e !== null)))
        for (var a = 0; a < e.length; a++) {
          var u = e[a], n = u.getSnapshot;
          u = u.value;
          try {
            if (!gl(n(), u)) return !1;
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
    l &= ~Li, l &= ~na, t.suspendedLanes |= l, t.pingedLanes &= ~l, a && (t.warmLanes |= l), a = t.expirationTimes;
    for (var u = l; 0 < u; ) {
      var n = 31 - ul(u), c = 1 << n;
      a[n] = -1, u &= ~c;
    }
    e !== 0 && fl(t, e, l);
  }
  function Vn() {
    return (ht & 6) === 0 ? (Hu(0), !1) : !0;
  }
  function $i() {
    if (et !== null) {
      if (yt === 0)
        var t = et.return;
      else
        t = et, ce = Ie = null, di(t), qa = null, Eu = 0, t = et;
      for (; t !== null; )
        eo(t.alternate, t), t = t.return;
      et = null;
    }
  }
  function Va(t, l) {
    var e = t.timeoutHandle;
    e !== -1 && (t.timeoutHandle = -1, A0(e)), e = t.cancelPendingCommit, e !== null && (t.cancelPendingCommit = null, e()), $i(), Tt = t, et = e = ae(t.current, null), ct = l, yt = 0, pl = null, Re = !1, Xa = E(t, l), Zi = !1, Qa = Hl = Li = na = _e = Ut = 0, hl = Uu = null, Vi = !1, (l & 8) !== 0 && (l |= l & 32);
    var a = t.entangledLanes;
    if (a !== 0)
      for (t = t.entanglements, a &= l; 0 < a; ) {
        var u = 31 - ul(a), n = 1 << u;
        l |= t[u], a &= ~n;
      }
    return ye = l, dn(), e;
  }
  function Eo(t, l) {
    tt = null, p.H = Rn, l === mu || l === pn ? (l = Qs(), yt = 3) : l === Ys ? (l = Qs(), yt = 4) : yt = l === Zr ? 8 : l !== null && typeof l == "object" && typeof l.then == "function" ? 6 : 1, pl = l, et === null && (Ut = 1, qn(
      t,
      Ol(l, t.current)
    ));
  }
  function zo() {
    var t = p.H;
    return p.H = Rn, t === null ? Rn : t;
  }
  function Mo() {
    var t = p.A;
    return p.A = u0, t;
  }
  function Wi() {
    Ut = 4, Re || (ct & 4194048) !== ct && Cl.current !== null || (Xa = !0), (_e & 134217727) === 0 && (na & 134217727) === 0 || Tt === null || He(
      Tt,
      ct,
      Hl,
      !1
    );
  }
  function Fi(t, l, e) {
    var a = ht;
    ht |= 2;
    var u = zo(), n = Mo();
    (Tt !== t || ct !== l) && (Ln = null, Va(t, l)), l = !1;
    var c = Ut;
    t: do
      try {
        if (yt !== 0 && et !== null) {
          var f = et, s = pl;
          switch (yt) {
            case 8:
              $i(), c = 6;
              break t;
            case 3:
            case 2:
            case 9:
            case 6:
              Cl.current === null && (l = !0);
              var m = yt;
              if (yt = 0, pl = null, wa(t, f, s, m), e && Xa) {
                c = 0;
                break t;
              }
              break;
            default:
              m = yt, yt = 0, pl = null, wa(t, f, s, m);
          }
        }
        i0(), c = Ut;
        break;
      } catch (j) {
        Eo(t, j);
      }
    while (!0);
    return l && t.shellSuspendCounter++, ce = Ie = null, ht = a, p.H = u, p.A = n, et === null && (Tt = null, ct = 0, dn()), c;
  }
  function i0() {
    for (; et !== null; ) No(et);
  }
  function f0(t, l) {
    var e = ht;
    ht |= 2;
    var a = zo(), u = Mo();
    Tt !== t || ct !== l ? (Ln = null, Zn = al() + 500, Va(t, l)) : Xa = E(
      t,
      l
    );
    t: do
      try {
        if (yt !== 0 && et !== null) {
          l = et;
          var n = pl;
          l: switch (yt) {
            case 1:
              yt = 0, pl = null, wa(t, l, n, 1);
              break;
            case 2:
            case 9:
              if (Gs(n)) {
                yt = 0, pl = null, Do(l);
                break;
              }
              l = function() {
                yt !== 2 && yt !== 9 || Tt !== t || (yt = 7), Pl(t);
              }, n.then(l, l);
              break t;
            case 3:
              yt = 7;
              break t;
            case 4:
              yt = 5;
              break t;
            case 7:
              Gs(n) ? (yt = 0, pl = null, Do(l)) : (yt = 0, pl = null, wa(t, l, n, 7));
              break;
            case 5:
              var c = null;
              switch (et.tag) {
                case 26:
                  c = et.memoizedState;
                case 5:
                case 27:
                  var f = et;
                  if (!c || rd(c)) {
                    yt = 0, pl = null;
                    var s = f.sibling;
                    if (s !== null) et = s;
                    else {
                      var m = f.return;
                      m !== null ? (et = m, wn(m)) : et = null;
                    }
                    break l;
                  }
              }
              yt = 0, pl = null, wa(t, l, n, 5);
              break;
            case 6:
              yt = 0, pl = null, wa(t, l, n, 6);
              break;
            case 8:
              $i(), Ut = 6;
              break t;
            default:
              throw Error(r(462));
          }
        }
        s0();
        break;
      } catch (j) {
        Eo(t, j);
      }
    while (!0);
    return ce = Ie = null, p.H = a, p.A = u, ht = e, et !== null ? 0 : (Tt = null, ct = 0, dn(), Ut);
  }
  function s0() {
    for (; et !== null && !Wu(); )
      No(et);
  }
  function No(t) {
    var l = to(t.alternate, t, ye);
    t.memoizedProps = t.pendingProps, l === null ? wn(t) : et = l;
  }
  function Do(t) {
    var l = t, e = l.alternate;
    switch (l.tag) {
      case 15:
      case 0:
        l = kr(
          e,
          l,
          l.pendingProps,
          l.type,
          void 0,
          ct
        );
        break;
      case 11:
        l = kr(
          e,
          l,
          l.pendingProps,
          l.type.render,
          l.ref,
          ct
        );
        break;
      case 5:
        di(l);
      default:
        eo(e, l), l = et = Ds(l, ye), l = to(e, l, ye);
    }
    t.memoizedProps = t.pendingProps, l === null ? wn(t) : et = l;
  }
  function wa(t, l, e, a) {
    ce = Ie = null, di(l), qa = null, Eu = 0;
    var u = l.return;
    try {
      if (Ph(
        t,
        u,
        l,
        e,
        ct
      )) {
        Ut = 1, qn(
          t,
          Ol(e, t.current)
        ), et = null;
        return;
      }
    } catch (n) {
      if (u !== null) throw et = u, n;
      Ut = 1, qn(
        t,
        Ol(e, t.current)
      ), et = null;
      return;
    }
    l.flags & 32768 ? (dt || a === 1 ? t = !0 : Xa || (ct & 536870912) !== 0 ? t = !1 : (Re = t = !0, (a === 2 || a === 9 || a === 3 || a === 6) && (a = Cl.current, a !== null && a.tag === 13 && (a.flags |= 16384))), Oo(l, t)) : wn(l);
  }
  function wn(t) {
    var l = t;
    do {
      if ((l.flags & 32768) !== 0) {
        Oo(
          l,
          Re
        );
        return;
      }
      t = l.return;
      var e = t0(
        l.alternate,
        l,
        ye
      );
      if (e !== null) {
        et = e;
        return;
      }
      if (l = l.sibling, l !== null) {
        et = l;
        return;
      }
      et = l = t;
    } while (l !== null);
    Ut === 0 && (Ut = 5);
  }
  function Oo(t, l) {
    do {
      var e = l0(t.alternate, t);
      if (e !== null) {
        e.flags &= 32767, et = e;
        return;
      }
      if (e = t.return, e !== null && (e.flags |= 32768, e.subtreeFlags = 0, e.deletions = null), !l && (t = t.sibling, t !== null)) {
        et = t;
        return;
      }
      et = t = e;
    } while (t !== null);
    Ut = 6, et = null;
  }
  function Ro(t, l, e, a, u, n, c, f, s) {
    t.cancelPendingCommit = null;
    do
      Kn();
    while (Ft !== 0);
    if ((ht & 6) !== 0) throw Error(r(327));
    if (l !== null) {
      if (l === t.current) throw Error(r(177));
      if (n = l.lanes | l.childLanes, n |= Qc, oa(
        t,
        e,
        n,
        c,
        f,
        s
      ), t === Tt && (et = Tt = null, ct = 0), Za = l, Ce = t, La = e, Ki = n, Ji = u, po = a, (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? (t.callbackNode = null, t.callbackPriority = 0, h0(Ll, function() {
        return qo(), null;
      })) : (t.callbackNode = null, t.callbackPriority = 0), a = (l.flags & 13878) !== 0, (l.subtreeFlags & 13878) !== 0 || a) {
        a = p.T, p.T = null, u = C.p, C.p = 2, c = ht, ht |= 4;
        try {
          e0(t, l, e);
        } finally {
          ht = c, C.p = u, p.T = a;
        }
      }
      Ft = 1, _o(), Uo(), Co();
    }
  }
  function _o() {
    if (Ft === 1) {
      Ft = 0;
      var t = Ce, l = Za, e = (l.flags & 13878) !== 0;
      if ((l.subtreeFlags & 13878) !== 0 || e) {
        e = p.T, p.T = null;
        var a = C.p;
        C.p = 2;
        var u = ht;
        ht |= 4;
        try {
          yo(l, t);
          var n = rf, c = bs(t.containerInfo), f = n.focusedElem, s = n.selectionRange;
          if (c !== f && f && f.ownerDocument && Ss(
            f.ownerDocument.documentElement,
            f
          )) {
            if (s !== null && qc(f)) {
              var m = s.start, j = s.end;
              if (j === void 0 && (j = m), "selectionStart" in f)
                f.selectionStart = m, f.selectionEnd = Math.min(
                  j,
                  f.value.length
                );
              else {
                var A = f.ownerDocument || document, g = A && A.defaultView || window;
                if (g.getSelection) {
                  var S = g.getSelection(), k = f.textContent.length, K = Math.min(s.start, k), gt = s.end === void 0 ? K : Math.min(s.end, k);
                  !S.extend && K > gt && (c = gt, gt = K, K = c);
                  var y = gs(
                    f,
                    K
                  ), h = gs(
                    f,
                    gt
                  );
                  if (y && h && (S.rangeCount !== 1 || S.anchorNode !== y.node || S.anchorOffset !== y.offset || S.focusNode !== h.node || S.focusOffset !== h.offset)) {
                    var v = A.createRange();
                    v.setStart(y.node, y.offset), S.removeAllRanges(), K > gt ? (S.addRange(v), S.extend(h.node, h.offset)) : (v.setEnd(h.node, h.offset), S.addRange(v));
                  }
                }
              }
            }
            for (A = [], S = f; S = S.parentNode; )
              S.nodeType === 1 && A.push({
                element: S,
                left: S.scrollLeft,
                top: S.scrollTop
              });
            for (typeof f.focus == "function" && f.focus(), f = 0; f < A.length; f++) {
              var T = A[f];
              T.element.scrollLeft = T.left, T.element.scrollTop = T.top;
            }
          }
          uc = !!sf, rf = sf = null;
        } finally {
          ht = u, C.p = a, p.T = e;
        }
      }
      t.current = l, Ft = 2;
    }
  }
  function Uo() {
    if (Ft === 2) {
      Ft = 0;
      var t = Ce, l = Za, e = (l.flags & 8772) !== 0;
      if ((l.subtreeFlags & 8772) !== 0 || e) {
        e = p.T, p.T = null;
        var a = C.p;
        C.p = 2;
        var u = ht;
        ht |= 4;
        try {
          so(t, l.alternate, l);
        } finally {
          ht = u, C.p = a, p.T = e;
        }
      }
      Ft = 3;
    }
  }
  function Co() {
    if (Ft === 4 || Ft === 3) {
      Ft = 0, ml();
      var t = Ce, l = Za, e = La, a = po;
      (l.subtreeFlags & 10256) !== 0 || (l.flags & 10256) !== 0 ? Ft = 5 : (Ft = 0, Za = Ce = null, Ho(t, t.pendingLanes));
      var u = t.pendingLanes;
      if (u === 0 && (Ue = null), Le(e), l = l.stateNode, nt && typeof nt.onCommitFiberRoot == "function")
        try {
          nt.onCommitFiberRoot(
            zl,
            l,
            void 0,
            (l.current.flags & 128) === 128
          );
        } catch {
        }
      if (a !== null) {
        l = p.T, u = C.p, C.p = 2, p.T = null;
        try {
          for (var n = t.onRecoverableError, c = 0; c < a.length; c++) {
            var f = a[c];
            n(f.value, {
              componentStack: f.stack
            });
          }
        } finally {
          p.T = l, C.p = u;
        }
      }
      (La & 3) !== 0 && Kn(), Pl(t), u = t.pendingLanes, (e & 4194090) !== 0 && (u & 42) !== 0 ? t === ki ? Cu++ : (Cu = 0, ki = t) : Cu = 0, Hu(0);
    }
  }
  function Ho(t, l) {
    (t.pooledCacheLanes &= l) === 0 && (l = t.pooledCache, l != null && (t.pooledCache = null, yu(l)));
  }
  function Kn(t) {
    return _o(), Uo(), Co(), qo();
  }
  function qo() {
    if (Ft !== 5) return !1;
    var t = Ce, l = Ki;
    Ki = 0;
    var e = Le(La), a = p.T, u = C.p;
    try {
      C.p = 32 > e ? 32 : e, p.T = null, e = Ji, Ji = null;
      var n = Ce, c = La;
      if (Ft = 0, Za = Ce = null, La = 0, (ht & 6) !== 0) throw Error(r(331));
      var f = ht;
      if (ht |= 4, bo(n.current), mo(
        n,
        n.current,
        c,
        e
      ), ht = f, Hu(0, !1), nt && typeof nt.onPostCommitFiberRoot == "function")
        try {
          nt.onPostCommitFiberRoot(zl, n);
        } catch {
        }
      return !0;
    } finally {
      C.p = u, p.T = a, Ho(t, l);
    }
  }
  function Bo(t, l, e) {
    l = Ol(e, l), l = zi(t.stateNode, l, 2), t = Ae(t, l, 2), t !== null && (Ct(t, 2), Pl(t));
  }
  function St(t, l, e) {
    if (t.tag === 3)
      Bo(t, t, e);
    else
      for (; l !== null; ) {
        if (l.tag === 3) {
          Bo(
            l,
            t,
            e
          );
          break;
        } else if (l.tag === 1) {
          var a = l.stateNode;
          if (typeof l.type.getDerivedStateFromError == "function" || typeof a.componentDidCatch == "function" && (Ue === null || !Ue.has(a))) {
            t = Ol(e, t), e = Xr(2), a = Ae(l, e, 2), a !== null && (Qr(
              e,
              a,
              l,
              t
            ), Ct(a, 2), Pl(a));
            break;
          }
        }
        l = l.return;
      }
  }
  function Pi(t, l, e) {
    var a = t.pingCache;
    if (a === null) {
      a = t.pingCache = new n0();
      var u = /* @__PURE__ */ new Set();
      a.set(l, u);
    } else
      u = a.get(l), u === void 0 && (u = /* @__PURE__ */ new Set(), a.set(l, u));
    u.has(e) || (Zi = !0, u.add(e), t = r0.bind(null, t, l, e), l.then(t, t));
  }
  function r0(t, l, e) {
    var a = t.pingCache;
    a !== null && a.delete(l), t.pingedLanes |= t.suspendedLanes & e, t.warmLanes &= ~e, Tt === t && (ct & e) === e && (Ut === 4 || Ut === 3 && (ct & 62914560) === ct && 300 > al() - wi ? (ht & 2) === 0 && Va(t, 0) : Li |= e, Qa === ct && (Qa = 0)), Pl(t);
  }
  function Yo(t, l) {
    l === 0 && (l = jt()), t = za(t, l), t !== null && (Ct(t, l), Pl(t));
  }
  function o0(t) {
    var l = t.memoizedState, e = 0;
    l !== null && (e = l.retryLane), Yo(t, e);
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
    a !== null && a.delete(l), Yo(t, e);
  }
  function h0(t, l) {
    return fa(t, l);
  }
  var Jn = null, Ka = null, Ii = !1, kn = !1, tf = !1, ca = 0;
  function Pl(t) {
    t !== Ka && t.next === null && (Ka === null ? Jn = Ka = t : Ka = Ka.next = t), kn = !0, Ii || (Ii = !0, v0());
  }
  function Hu(t, l) {
    if (!tf && kn) {
      tf = !0;
      do
        for (var e = !1, a = Jn; a !== null; ) {
          if (t !== 0) {
            var u = a.pendingLanes;
            if (u === 0) var n = 0;
            else {
              var c = a.suspendedLanes, f = a.pingedLanes;
              n = (1 << 31 - ul(42 | t) + 1) - 1, n &= u & ~(c & ~f), n = n & 201326741 ? n & 201326741 | 1 : n ? n | 2 : 0;
            }
            n !== 0 && (e = !0, Zo(a, n));
          } else
            n = ct, n = wl(
              a,
              a === Tt ? n : 0,
              a.cancelPendingCommit !== null || a.timeoutHandle !== -1
            ), (n & 3) === 0 || E(a, n) || (e = !0, Zo(a, n));
          a = a.next;
        }
      while (e);
      tf = !1;
    }
  }
  function y0() {
    Go();
  }
  function Go() {
    kn = Ii = !1;
    var t = 0;
    ca !== 0 && (T0() && (t = ca), ca = 0);
    for (var l = al(), e = null, a = Jn; a !== null; ) {
      var u = a.next, n = Xo(a, l);
      n === 0 ? (a.next = null, e === null ? Jn = u : e.next = u, u === null && (Ka = e)) : (e = a, (t !== 0 || (n & 3) !== 0) && (kn = !0)), a = u;
    }
    Hu(t);
  }
  function Xo(t, l) {
    for (var e = t.suspendedLanes, a = t.pingedLanes, u = t.expirationTimes, n = t.pendingLanes & -62914561; 0 < n; ) {
      var c = 31 - ul(n), f = 1 << c, s = u[c];
      s === -1 ? ((f & e) === 0 || (f & a) !== 0) && (u[c] = Q(f, l)) : s <= l && (t.expiredLanes |= f), n &= ~f;
    }
    if (l = Tt, e = ct, e = wl(
      t,
      t === l ? e : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a = t.callbackNode, e === 0 || t === l && (yt === 2 || yt === 9) || t.cancelPendingCommit !== null)
      return a !== null && a !== null && Fa(a), t.callbackNode = null, t.callbackPriority = 0;
    if ((e & 3) === 0 || E(t, e)) {
      if (l = e & -e, l === t.callbackPriority) return l;
      switch (a !== null && Fa(a), Le(e)) {
        case 2:
        case 8:
          e = cl;
          break;
        case 32:
          e = Ll;
          break;
        case 268435456:
          e = Rt;
          break;
        default:
          e = Ll;
      }
      return a = Qo.bind(null, t), e = fa(e, a), t.callbackPriority = l, t.callbackNode = e, l;
    }
    return a !== null && a !== null && Fa(a), t.callbackPriority = 2, t.callbackNode = null, 2;
  }
  function Qo(t, l) {
    if (Ft !== 0 && Ft !== 5)
      return t.callbackNode = null, t.callbackPriority = 0, null;
    var e = t.callbackNode;
    if (Kn() && t.callbackNode !== e)
      return null;
    var a = ct;
    return a = wl(
      t,
      t === Tt ? a : 0,
      t.cancelPendingCommit !== null || t.timeoutHandle !== -1
    ), a === 0 ? null : (To(t, a, l), Xo(t, al()), t.callbackNode != null && t.callbackNode === e ? Qo.bind(null, t) : null);
  }
  function Zo(t, l) {
    if (Kn()) return null;
    To(t, l, !0);
  }
  function v0() {
    E0(function() {
      (ht & 6) !== 0 ? fa(
        Zl,
        y0
      ) : Go();
    });
  }
  function lf() {
    return ca === 0 && (ca = F()), ca;
  }
  function Lo(t) {
    return t == null || typeof t == "symbol" || typeof t == "boolean" ? null : typeof t == "function" ? t : un("" + t);
  }
  function Vo(t, l) {
    var e = l.ownerDocument.createElement("input");
    return e.name = l.name, e.value = l.value, t.id && e.setAttribute("form", t.id), l.parentNode.insertBefore(e, l), t = new FormData(t), e.parentNode.removeChild(e), t;
  }
  function m0(t, l, e, a, u) {
    if (l === "submit" && e && e.stateNode === u) {
      var n = Lo(
        (u[sl] || null).action
      ), c = a.submitter;
      c && (l = (l = c[sl] || null) ? Lo(l.formAction) : c.getAttribute("formAction"), l !== null && (n = l, c = null));
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
                if (ca !== 0) {
                  var s = c ? Vo(u, c) : new FormData(u);
                  pi(
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
                typeof n == "function" && (f.preventDefault(), s = c ? Vo(u, c) : new FormData(u), pi(
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
  for (var ef = 0; ef < Xc.length; ef++) {
    var af = Xc[ef], g0 = af.toLowerCase(), S0 = af[0].toUpperCase() + af.slice(1);
    Gl(
      g0,
      "on" + S0
    );
  }
  Gl(js, "onAnimationEnd"), Gl(Ts, "onAnimationIteration"), Gl(As, "onAnimationStart"), Gl("dblclick", "onDoubleClick"), Gl("focusin", "onFocus"), Gl("focusout", "onBlur"), Gl(Hh, "onTransitionRun"), Gl(qh, "onTransitionStart"), Gl(Bh, "onTransitionCancel"), Gl(Es, "onTransitionEnd"), ma("onMouseEnter", ["mouseout", "mouseover"]), ma("onMouseLeave", ["mouseout", "mouseover"]), ma("onPointerEnter", ["pointerout", "pointerover"]), ma("onPointerLeave", ["pointerout", "pointerover"]), Ve(
    "onChange",
    "change click focusin focusout input keydown keyup selectionchange".split(" ")
  ), Ve(
    "onSelect",
    "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(
      " "
    )
  ), Ve("onBeforeInput", [
    "compositionend",
    "keypress",
    "textInput",
    "paste"
  ]), Ve(
    "onCompositionEnd",
    "compositionend focusout keydown keypress keyup mousedown".split(" ")
  ), Ve(
    "onCompositionStart",
    "compositionstart focusout keydown keypress keyup mousedown".split(" ")
  ), Ve(
    "onCompositionUpdate",
    "compositionupdate focusout keydown keypress keyup mousedown".split(" ")
  );
  var qu = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(
    " "
  ), b0 = new Set(
    "beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(qu)
  );
  function wo(t, l) {
    l = (l & 4) !== 0;
    for (var e = 0; e < t.length; e++) {
      var a = t[e], u = a.event;
      a = a.listeners;
      t: {
        var n = void 0;
        if (l)
          for (var c = a.length - 1; 0 <= c; c--) {
            var f = a[c], s = f.instance, m = f.currentTarget;
            if (f = f.listener, s !== n && u.isPropagationStopped())
              break t;
            n = f, u.currentTarget = m;
            try {
              n(u);
            } catch (j) {
              Hn(j);
            }
            u.currentTarget = null, n = s;
          }
        else
          for (c = 0; c < a.length; c++) {
            if (f = a[c], s = f.instance, m = f.currentTarget, f = f.listener, s !== n && u.isPropagationStopped())
              break t;
            n = f, u.currentTarget = m;
            try {
              n(u);
            } catch (j) {
              Hn(j);
            }
            u.currentTarget = null, n = s;
          }
      }
    }
  }
  function at(t, l) {
    var e = l[mc];
    e === void 0 && (e = l[mc] = /* @__PURE__ */ new Set());
    var a = t + "__bubble";
    e.has(a) || (Ko(l, t, 2, !1), e.add(a));
  }
  function uf(t, l, e) {
    var a = 0;
    l && (a |= 4), Ko(
      e,
      t,
      a,
      l
    );
  }
  var $n = "_reactListening" + Math.random().toString(36).slice(2);
  function nf(t) {
    if (!t[$n]) {
      t[$n] = !0, Yf.forEach(function(e) {
        e !== "selectionchange" && (b0.has(e) || uf(e, !1, t), uf(e, !0, t));
      });
      var l = t.nodeType === 9 ? t : t.ownerDocument;
      l === null || l[$n] || (l[$n] = !0, uf("selectionchange", !1, l));
    }
  }
  function Ko(t, l, e, a) {
    switch (md(l)) {
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
    ), u = void 0, !Mc || l !== "touchstart" && l !== "touchmove" && l !== "wheel" || (u = !0), a ? u !== void 0 ? t.addEventListener(l, e, {
      capture: !0,
      passive: u
    }) : t.addEventListener(l, e, !0) : u !== void 0 ? t.addEventListener(l, e, {
      passive: u
    }) : t.addEventListener(l, e, !1);
  }
  function cf(t, l, e, a, u) {
    var n = a;
    if ((l & 1) === 0 && (l & 2) === 0 && a !== null)
      t: for (; ; ) {
        if (a === null) return;
        var c = a.tag;
        if (c === 3 || c === 4) {
          var f = a.stateNode.containerInfo;
          if (f === u) break;
          if (c === 4)
            for (c = a.return; c !== null; ) {
              var s = c.tag;
              if ((s === 3 || s === 4) && c.stateNode.containerInfo === u)
                return;
              c = c.return;
            }
          for (; f !== null; ) {
            if (c = ha(f), c === null) return;
            if (s = c.tag, s === 5 || s === 6 || s === 26 || s === 27) {
              a = n = c;
              continue t;
            }
            f = f.parentNode;
          }
        }
        a = a.return;
      }
    Pf(function() {
      var m = n, j = Ec(e), A = [];
      t: {
        var g = zs.get(t);
        if (g !== void 0) {
          var S = sn, k = t;
          switch (t) {
            case "keypress":
              if (cn(e) === 0) break t;
            case "keydown":
            case "keyup":
              S = hh;
              break;
            case "focusin":
              k = "focus", S = Rc;
              break;
            case "focusout":
              k = "blur", S = Rc;
              break;
            case "beforeblur":
            case "afterblur":
              S = Rc;
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
              S = ls;
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
            case js:
            case Ts:
            case As:
              S = uh;
              break;
            case Es:
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
              S = ch;
              break;
            case "gotpointercapture":
            case "lostpointercapture":
            case "pointercancel":
            case "pointerdown":
            case "pointermove":
            case "pointerout":
            case "pointerover":
            case "pointerup":
              S = as;
              break;
            case "toggle":
            case "beforetoggle":
              S = jh;
          }
          var K = (l & 4) !== 0, gt = !K && (t === "scroll" || t === "scrollend"), y = K ? g !== null ? g + "Capture" : null : g;
          K = [];
          for (var h = m, v; h !== null; ) {
            var T = h;
            if (v = T.stateNode, T = T.tag, T !== 5 && T !== 26 && T !== 27 || v === null || y === null || (T = lu(h, y), T != null && K.push(
              Bu(h, T, v)
            )), gt) break;
            h = h.return;
          }
          0 < K.length && (g = new S(
            g,
            k,
            null,
            e,
            j
          ), A.push({ event: g, listeners: K }));
        }
      }
      if ((l & 7) === 0) {
        t: {
          if (g = t === "mouseover" || t === "pointerover", S = t === "mouseout" || t === "pointerout", g && e !== Ac && (k = e.relatedTarget || e.fromElement) && (ha(k) || k[da]))
            break t;
          if ((S || g) && (g = j.window === j ? j : (g = j.ownerDocument) ? g.defaultView || g.parentWindow : window, S ? (k = e.relatedTarget || e.toElement, S = m, k = k ? ha(k) : null, k !== null && (gt = D(k), K = k.tag, k !== gt || K !== 5 && K !== 27 && K !== 6) && (k = null)) : (S = null, k = m), S !== k)) {
            if (K = ls, T = "onMouseLeave", y = "onMouseEnter", h = "mouse", (t === "pointerout" || t === "pointerover") && (K = as, T = "onPointerLeave", y = "onPointerEnter", h = "pointer"), gt = S == null ? g : tu(S), v = k == null ? g : tu(k), g = new K(
              T,
              h + "leave",
              S,
              e,
              j
            ), g.target = gt, g.relatedTarget = v, T = null, ha(j) === m && (K = new K(
              y,
              h + "enter",
              k,
              e,
              j
            ), K.target = v, K.relatedTarget = gt, T = K), gt = T, S && k)
              l: {
                for (K = S, y = k, h = 0, v = K; v; v = Ja(v))
                  h++;
                for (v = 0, T = y; T; T = Ja(T))
                  v++;
                for (; 0 < h - v; )
                  K = Ja(K), h--;
                for (; 0 < v - h; )
                  y = Ja(y), v--;
                for (; h--; ) {
                  if (K === y || y !== null && K === y.alternate)
                    break l;
                  K = Ja(K), y = Ja(y);
                }
                K = null;
              }
            else K = null;
            S !== null && Jo(
              A,
              g,
              S,
              K,
              !1
            ), k !== null && gt !== null && Jo(
              A,
              gt,
              k,
              K,
              !0
            );
          }
        }
        t: {
          if (g = m ? tu(m) : window, S = g.nodeName && g.nodeName.toLowerCase(), S === "select" || S === "input" && g.type === "file")
            var G = os;
          else if (ss(g))
            if (ds)
              G = _h;
            else {
              G = Oh;
              var lt = Dh;
            }
          else
            S = g.nodeName, !S || S.toLowerCase() !== "input" || g.type !== "checkbox" && g.type !== "radio" ? m && Tc(m.elementType) && (G = os) : G = Rh;
          if (G && (G = G(t, m))) {
            rs(
              A,
              G,
              e,
              j
            );
            break t;
          }
          lt && lt(t, g, m), t === "focusout" && m && g.type === "number" && m.memoizedProps.value != null && jc(g, "number", g.value);
        }
        switch (lt = m ? tu(m) : window, t) {
          case "focusin":
            (ss(lt) || lt.contentEditable === "true") && (Ta = lt, Bc = m, su = null);
            break;
          case "focusout":
            su = Bc = Ta = null;
            break;
          case "mousedown":
            Yc = !0;
            break;
          case "contextmenu":
          case "mouseup":
          case "dragend":
            Yc = !1, xs(A, e, j);
            break;
          case "selectionchange":
            if (Ch) break;
          case "keydown":
          case "keyup":
            xs(A, e, j);
        }
        var L;
        if (Uc)
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
          ja ? is(t, e) && (J = "onCompositionEnd") : t === "keydown" && e.keyCode === 229 && (J = "onCompositionStart");
        J && (us && e.locale !== "ko" && (ja || J !== "onCompositionStart" ? J === "onCompositionEnd" && ja && (L = If()) : (xe = j, Nc = "value" in xe ? xe.value : xe.textContent, ja = !0)), lt = Wn(m, J), 0 < lt.length && (J = new es(
          J,
          t,
          null,
          e,
          j
        ), A.push({ event: J, listeners: lt }), L ? J.data = L : (L = fs(e), L !== null && (J.data = L)))), (L = Ah ? Eh(t, e) : zh(t, e)) && (J = Wn(m, "onBeforeInput"), 0 < J.length && (lt = new es(
          "onBeforeInput",
          "beforeinput",
          null,
          e,
          j
        ), A.push({
          event: lt,
          listeners: J
        }), lt.data = L)), m0(
          A,
          t,
          m,
          e,
          j
        );
      }
      wo(A, l);
    });
  }
  function Bu(t, l, e) {
    return {
      instance: t,
      listener: l,
      currentTarget: e
    };
  }
  function Wn(t, l) {
    for (var e = l + "Capture", a = []; t !== null; ) {
      var u = t, n = u.stateNode;
      if (u = u.tag, u !== 5 && u !== 26 && u !== 27 || n === null || (u = lu(t, e), u != null && a.unshift(
        Bu(t, u, n)
      ), u = lu(t, l), u != null && a.push(
        Bu(t, u, n)
      )), t.tag === 3) return a;
      t = t.return;
    }
    return [];
  }
  function Ja(t) {
    if (t === null) return null;
    do
      t = t.return;
    while (t && t.tag !== 5 && t.tag !== 27);
    return t || null;
  }
  function Jo(t, l, e, a, u) {
    for (var n = l._reactName, c = []; e !== null && e !== a; ) {
      var f = e, s = f.alternate, m = f.stateNode;
      if (f = f.tag, s !== null && s === a) break;
      f !== 5 && f !== 26 && f !== 27 || m === null || (s = m, u ? (m = lu(e, n), m != null && c.unshift(
        Bu(e, m, s)
      )) : u || (m = lu(e, n), m != null && c.push(
        Bu(e, m, s)
      ))), e = e.return;
    }
    c.length !== 0 && t.push({ event: l, listeners: c });
  }
  var x0 = /\r\n?/g, p0 = /\u0000|\uFFFD/g;
  function ko(t) {
    return (typeof t == "string" ? t : "" + t).replace(x0, `
`).replace(p0, "");
  }
  function $o(t, l) {
    return l = ko(l), ko(t) === l;
  }
  function Fn() {
  }
  function mt(t, l, e, a, u, n) {
    switch (e) {
      case "children":
        typeof a == "string" ? l === "body" || l === "textarea" && a === "" || ba(t, a) : (typeof a == "number" || typeof a == "bigint") && l !== "body" && ba(t, "" + a);
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
        Wf(t, a, n);
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
          typeof n == "function" && (e === "formAction" ? (l !== "input" && mt(t, l, "name", u.name, u, null), mt(
            t,
            l,
            "formEncType",
            u.formEncType,
            u,
            null
          ), mt(
            t,
            l,
            "formMethod",
            u.formMethod,
            u,
            null
          ), mt(
            t,
            l,
            "formTarget",
            u.formTarget,
            u,
            null
          )) : (mt(t, l, "encType", u.encType, u, null), mt(t, l, "method", u.method, u, null), mt(t, l, "target", u.target, u, null)));
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
        a != null && at("scroll", t);
        break;
      case "onScrollEnd":
        a != null && at("scrollend", t);
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
        at("beforetoggle", t), at("toggle", t), tn(t, "popover", a);
        break;
      case "xlinkActuate":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:actuate",
          a
        );
        break;
      case "xlinkArcrole":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:arcrole",
          a
        );
        break;
      case "xlinkRole":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:role",
          a
        );
        break;
      case "xlinkShow":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:show",
          a
        );
        break;
      case "xlinkTitle":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:title",
          a
        );
        break;
      case "xlinkType":
        le(
          t,
          "http://www.w3.org/1999/xlink",
          "xlink:type",
          a
        );
        break;
      case "xmlBase":
        le(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:base",
          a
        );
        break;
      case "xmlLang":
        le(
          t,
          "http://www.w3.org/XML/1998/namespace",
          "xml:lang",
          a
        );
        break;
      case "xmlSpace":
        le(
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
  function ff(t, l, e, a, u, n) {
    switch (e) {
      case "style":
        Wf(t, a, n);
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
        typeof a == "string" ? ba(t, a) : (typeof a == "number" || typeof a == "bigint") && ba(t, "" + a);
        break;
      case "onScroll":
        a != null && at("scroll", t);
        break;
      case "onScrollEnd":
        a != null && at("scrollend", t);
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
        if (!Gf.hasOwnProperty(e))
          t: {
            if (e[0] === "o" && e[1] === "n" && (u = e.endsWith("Capture"), l = e.slice(2, u ? e.length - 7 : void 0), n = t[sl] || null, n = n != null ? n[e] : null, typeof n == "function" && t.removeEventListener(l, n, u), typeof a == "function")) {
              typeof n != "function" && n !== null && (e in t ? t[e] = null : t.hasAttribute(e) && t.removeAttribute(e)), t.addEventListener(l, a, u);
              break t;
            }
            e in t ? t[e] = a : a === !0 ? t.setAttribute(e, "") : tn(t, e, a);
          }
    }
  }
  function Pt(t, l, e) {
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
        at("error", t), at("load", t);
        var a = !1, u = !1, n;
        for (n in e)
          if (e.hasOwnProperty(n)) {
            var c = e[n];
            if (c != null)
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
                  mt(t, l, n, c, e, null);
              }
          }
        u && mt(t, l, "srcSet", e.srcSet, e, null), a && mt(t, l, "src", e.src, e, null);
        return;
      case "input":
        at("invalid", t);
        var f = n = c = u = null, s = null, m = null;
        for (a in e)
          if (e.hasOwnProperty(a)) {
            var j = e[a];
            if (j != null)
              switch (a) {
                case "name":
                  u = j;
                  break;
                case "type":
                  c = j;
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
                    throw Error(r(137, l));
                  break;
                default:
                  mt(t, l, a, j, e, null);
              }
          }
        Kf(
          t,
          n,
          f,
          s,
          m,
          c,
          u,
          !1
        ), en(t);
        return;
      case "select":
        at("invalid", t), a = c = n = null;
        for (u in e)
          if (e.hasOwnProperty(u) && (f = e[u], f != null))
            switch (u) {
              case "value":
                n = f;
                break;
              case "defaultValue":
                c = f;
                break;
              case "multiple":
                a = f;
              default:
                mt(t, l, u, f, e, null);
            }
        l = n, e = c, t.multiple = !!a, l != null ? Sa(t, !!a, l, !1) : e != null && Sa(t, !!a, e, !0);
        return;
      case "textarea":
        at("invalid", t), n = u = a = null;
        for (c in e)
          if (e.hasOwnProperty(c) && (f = e[c], f != null))
            switch (c) {
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
                mt(t, l, c, f, e, null);
            }
        kf(t, a, u, n), en(t);
        return;
      case "option":
        for (s in e)
          e.hasOwnProperty(s) && (a = e[s], a != null) && (s === "selected" ? t.selected = a && typeof a != "function" && typeof a != "symbol" : mt(t, l, s, a, e, null));
        return;
      case "dialog":
        at("beforetoggle", t), at("toggle", t), at("cancel", t), at("close", t);
        break;
      case "iframe":
      case "object":
        at("load", t);
        break;
      case "video":
      case "audio":
        for (a = 0; a < qu.length; a++)
          at(qu[a], t);
        break;
      case "image":
        at("error", t), at("load", t);
        break;
      case "details":
        at("toggle", t);
        break;
      case "embed":
      case "source":
      case "link":
        at("error", t), at("load", t);
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
                mt(t, l, m, a, e, null);
            }
        return;
      default:
        if (Tc(l)) {
          for (j in e)
            e.hasOwnProperty(j) && (a = e[j], a !== void 0 && ff(
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
      e.hasOwnProperty(f) && (a = e[f], a != null && mt(t, l, f, a, e, null));
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
        var u = null, n = null, c = null, f = null, s = null, m = null, j = null;
        for (S in e) {
          var A = e[S];
          if (e.hasOwnProperty(S) && A != null)
            switch (S) {
              case "checked":
                break;
              case "value":
                break;
              case "defaultValue":
                s = A;
              default:
                a.hasOwnProperty(S) || mt(t, l, S, null, a, A);
            }
        }
        for (var g in a) {
          var S = a[g];
          if (A = e[g], a.hasOwnProperty(g) && (S != null || A != null))
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
                c = S;
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
                S !== A && mt(
                  t,
                  l,
                  g,
                  S,
                  a,
                  A
                );
            }
        }
        pc(
          t,
          c,
          f,
          s,
          m,
          j,
          n,
          u
        );
        return;
      case "select":
        S = c = f = g = null;
        for (n in e)
          if (s = e[n], e.hasOwnProperty(n) && s != null)
            switch (n) {
              case "value":
                break;
              case "multiple":
                S = s;
              default:
                a.hasOwnProperty(n) || mt(
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
                c = n;
              default:
                n !== s && mt(
                  t,
                  l,
                  u,
                  n,
                  a,
                  s
                );
            }
        l = f, e = c, a = S, g != null ? Sa(t, !!e, g, !1) : !!a != !!e && (l != null ? Sa(t, !!e, l, !0) : Sa(t, !!e, e ? [] : "", !1));
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
                mt(t, l, f, null, a, u);
            }
        for (c in a)
          if (u = a[c], n = e[c], a.hasOwnProperty(c) && (u != null || n != null))
            switch (c) {
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
                u !== n && mt(t, l, c, u, a, n);
            }
        Jf(t, g, S);
        return;
      case "option":
        for (var k in e)
          g = e[k], e.hasOwnProperty(k) && g != null && !a.hasOwnProperty(k) && (k === "selected" ? t.selected = !1 : mt(
            t,
            l,
            k,
            null,
            a,
            g
          ));
        for (s in a)
          g = a[s], S = e[s], a.hasOwnProperty(s) && g !== S && (g != null || S != null) && (s === "selected" ? t.selected = g && typeof g != "function" && typeof g != "symbol" : mt(
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
          g = e[K], e.hasOwnProperty(K) && g != null && !a.hasOwnProperty(K) && mt(t, l, K, null, a, g);
        for (m in a)
          if (g = a[m], S = e[m], a.hasOwnProperty(m) && g !== S && (g != null || S != null))
            switch (m) {
              case "children":
              case "dangerouslySetInnerHTML":
                if (g != null)
                  throw Error(r(137, l));
                break;
              default:
                mt(
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
        if (Tc(l)) {
          for (var gt in e)
            g = e[gt], e.hasOwnProperty(gt) && g !== void 0 && !a.hasOwnProperty(gt) && ff(
              t,
              l,
              gt,
              void 0,
              a,
              g
            );
          for (j in a)
            g = a[j], S = e[j], !a.hasOwnProperty(j) || g === S || g === void 0 && S === void 0 || ff(
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
      g = e[y], e.hasOwnProperty(y) && g != null && !a.hasOwnProperty(y) && mt(t, l, y, null, a, g);
    for (A in a)
      g = a[A], S = e[A], !a.hasOwnProperty(A) || g === S || g == null && S == null || mt(t, l, A, g, a, S);
  }
  var sf = null, rf = null;
  function Pn(t) {
    return t.nodeType === 9 ? t : t.ownerDocument;
  }
  function Wo(t) {
    switch (t) {
      case "http://www.w3.org/2000/svg":
        return 1;
      case "http://www.w3.org/1998/Math/MathML":
        return 2;
      default:
        return 0;
    }
  }
  function Fo(t, l) {
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
  function of(t, l) {
    return t === "textarea" || t === "noscript" || typeof l.children == "string" || typeof l.children == "number" || typeof l.children == "bigint" || typeof l.dangerouslySetInnerHTML == "object" && l.dangerouslySetInnerHTML !== null && l.dangerouslySetInnerHTML.__html != null;
  }
  var df = null;
  function T0() {
    var t = window.event;
    return t && t.type === "popstate" ? t === df ? !1 : (df = t, !0) : (df = null, !1);
  }
  var Po = typeof setTimeout == "function" ? setTimeout : void 0, A0 = typeof clearTimeout == "function" ? clearTimeout : void 0, Io = typeof Promise == "function" ? Promise : void 0, E0 = typeof queueMicrotask == "function" ? queueMicrotask : typeof Io < "u" ? function(t) {
    return Io.resolve(null).then(t).catch(z0);
  } : Po;
  function z0(t) {
    setTimeout(function() {
      throw t;
    });
  }
  function qe(t) {
    return t === "head";
  }
  function td(t, l) {
    var e = l, a = 0, u = 0;
    do {
      var n = e.nextSibling;
      if (t.removeChild(e), n && n.nodeType === 8)
        if (e = n.data, e === "/$") {
          if (0 < a && 8 > a) {
            e = a;
            var c = t.ownerDocument;
            if (e & 1 && Yu(c.documentElement), e & 2 && Yu(c.body), e & 4)
              for (e = c.head, Yu(e), c = e.firstChild; c; ) {
                var f = c.nextSibling, s = c.nodeName;
                c[Ia] || s === "SCRIPT" || s === "STYLE" || s === "LINK" && c.rel.toLowerCase() === "stylesheet" || e.removeChild(c), c = f;
              }
          }
          if (u === 0) {
            t.removeChild(n), Ku(l);
            return;
          }
          u--;
        } else
          e === "$" || e === "$?" || e === "$!" ? u++ : a = e.charCodeAt(0) - 48;
      else a = 0;
      e = n;
    } while (e);
    Ku(l);
  }
  function hf(t) {
    var l = t.firstChild;
    for (l && l.nodeType === 10 && (l = l.nextSibling); l; ) {
      var e = l;
      switch (l = l.nextSibling, e.nodeName) {
        case "HTML":
        case "HEAD":
        case "BODY":
          hf(e), gc(e);
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
        if (!t[Ia])
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
  function yf(t) {
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
  var vf = null;
  function ld(t) {
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
  function ed(t, l, e) {
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
  function Yu(t) {
    for (var l = t.attributes; l.length; )
      t.removeAttributeNode(l[0]);
    gc(t);
  }
  var ql = /* @__PURE__ */ new Map(), ad = /* @__PURE__ */ new Set();
  function In(t) {
    return typeof t.getRootNode == "function" ? t.getRootNode() : t.nodeType === 9 ? t : t.ownerDocument;
  }
  var ve = C.d;
  C.d = {
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
    var t = ve.f(), l = Vn();
    return t || l;
  }
  function R0(t) {
    var l = ya(t);
    l !== null && l.tag === 5 && l.type === "form" ? jr(l) : ve.r(t);
  }
  var ka = typeof document > "u" ? null : document;
  function ud(t, l, e) {
    var a = ka;
    if (a && typeof l == "string" && l) {
      var u = Dl(l);
      u = 'link[rel="' + t + '"][href="' + u + '"]', typeof e == "string" && (u += '[crossorigin="' + e + '"]'), ad.has(u) || (ad.add(u), t = { rel: t, crossOrigin: e, href: l }, a.querySelector(u) === null && (l = a.createElement("link"), Pt(l, "link", t), wt(l), a.head.appendChild(l)));
    }
  }
  function _0(t) {
    ve.D(t), ud("dns-prefetch", t, null);
  }
  function U0(t, l) {
    ve.C(t, l), ud("preconnect", t, l);
  }
  function C0(t, l, e) {
    ve.L(t, l, e);
    var a = ka;
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
          n = $a(t);
          break;
        case "script":
          n = Wa(t);
      }
      ql.has(n) || (t = B(
        {
          rel: "preload",
          href: l === "image" && e && e.imageSrcSet ? void 0 : t,
          as: l
        },
        e
      ), ql.set(n, t), a.querySelector(u) !== null || l === "style" && a.querySelector(Gu(n)) || l === "script" && a.querySelector(Xu(n)) || (l = a.createElement("link"), Pt(l, "link", t), wt(l), a.head.appendChild(l)));
    }
  }
  function H0(t, l) {
    ve.m(t, l);
    var e = ka;
    if (e && t) {
      var a = l && typeof l.as == "string" ? l.as : "script", u = 'link[rel="modulepreload"][as="' + Dl(a) + '"][href="' + Dl(t) + '"]', n = u;
      switch (a) {
        case "audioworklet":
        case "paintworklet":
        case "serviceworker":
        case "sharedworker":
        case "worker":
        case "script":
          n = Wa(t);
      }
      if (!ql.has(n) && (t = B({ rel: "modulepreload", href: t }, l), ql.set(n, t), e.querySelector(u) === null)) {
        switch (a) {
          case "audioworklet":
          case "paintworklet":
          case "serviceworker":
          case "sharedworker":
          case "worker":
          case "script":
            if (e.querySelector(Xu(n)))
              return;
        }
        a = e.createElement("link"), Pt(a, "link", t), wt(a), e.head.appendChild(a);
      }
    }
  }
  function q0(t, l, e) {
    ve.S(t, l, e);
    var a = ka;
    if (a && t) {
      var u = va(a).hoistableStyles, n = $a(t);
      l = l || "default";
      var c = u.get(n);
      if (!c) {
        var f = { loading: 0, preload: null };
        if (c = a.querySelector(
          Gu(n)
        ))
          f.loading = 5;
        else {
          t = B(
            { rel: "stylesheet", href: t, "data-precedence": l },
            e
          ), (e = ql.get(n)) && mf(t, e);
          var s = c = a.createElement("link");
          wt(s), Pt(s, "link", t), s._p = new Promise(function(m, j) {
            s.onload = m, s.onerror = j;
          }), s.addEventListener("load", function() {
            f.loading |= 1;
          }), s.addEventListener("error", function() {
            f.loading |= 2;
          }), f.loading |= 4, tc(c, l, a);
        }
        c = {
          type: "stylesheet",
          instance: c,
          count: 1,
          state: f
        }, u.set(n, c);
      }
    }
  }
  function B0(t, l) {
    ve.X(t, l);
    var e = ka;
    if (e && t) {
      var a = va(e).hoistableScripts, u = Wa(t), n = a.get(u);
      n || (n = e.querySelector(Xu(u)), n || (t = B({ src: t, async: !0 }, l), (l = ql.get(u)) && gf(t, l), n = e.createElement("script"), wt(n), Pt(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function Y0(t, l) {
    ve.M(t, l);
    var e = ka;
    if (e && t) {
      var a = va(e).hoistableScripts, u = Wa(t), n = a.get(u);
      n || (n = e.querySelector(Xu(u)), n || (t = B({ src: t, async: !0, type: "module" }, l), (l = ql.get(u)) && gf(t, l), n = e.createElement("script"), wt(n), Pt(n, "link", t), e.head.appendChild(n)), n = {
        type: "script",
        instance: n,
        count: 1,
        state: null
      }, a.set(u, n));
    }
  }
  function nd(t, l, e, a) {
    var u = (u = $.current) ? In(u) : null;
    if (!u) throw Error(r(446));
    switch (t) {
      case "meta":
      case "title":
        return null;
      case "style":
        return typeof e.precedence == "string" && typeof e.href == "string" ? (l = $a(e.href), e = va(
          u
        ).hoistableStyles, a = e.get(l), a || (a = {
          type: "style",
          instance: null,
          count: 0,
          state: null
        }, e.set(l, a)), a) : { type: "void", instance: null, count: 0, state: null };
      case "link":
        if (e.rel === "stylesheet" && typeof e.href == "string" && typeof e.precedence == "string") {
          t = $a(e.href);
          var n = va(
            u
          ).hoistableStyles, c = n.get(t);
          if (c || (u = u.ownerDocument || u, c = {
            type: "stylesheet",
            instance: null,
            count: 0,
            state: { loading: 0, preload: null }
          }, n.set(t, c), (n = u.querySelector(
            Gu(t)
          )) && !n._p && (c.instance = n, c.state.loading = 5), ql.has(t) || (e = {
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
            c.state
          ))), l && a === null)
            throw Error(r(528, ""));
          return c;
        }
        if (l && a !== null)
          throw Error(r(529, ""));
        return null;
      case "script":
        return l = e.async, e = e.src, typeof e == "string" && l && typeof l != "function" && typeof l != "symbol" ? (l = Wa(e), e = va(
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
  function $a(t) {
    return 'href="' + Dl(t) + '"';
  }
  function Gu(t) {
    return 'link[rel="stylesheet"][' + t + "]";
  }
  function cd(t) {
    return B({}, t, {
      "data-precedence": t.precedence,
      precedence: null
    });
  }
  function G0(t, l, e, a) {
    t.querySelector('link[rel="preload"][as="style"][' + l + "]") ? a.loading = 1 : (l = t.createElement("link"), a.preload = l, l.addEventListener("load", function() {
      return a.loading |= 1;
    }), l.addEventListener("error", function() {
      return a.loading |= 2;
    }), Pt(l, "link", e), wt(l), t.head.appendChild(l));
  }
  function Wa(t) {
    return '[src="' + Dl(t) + '"]';
  }
  function Xu(t) {
    return "script[async]" + t;
  }
  function id(t, l, e) {
    if (l.count++, l.instance === null)
      switch (l.type) {
        case "style":
          var a = t.querySelector(
            'style[data-href~="' + Dl(e.href) + '"]'
          );
          if (a)
            return l.instance = a, wt(a), a;
          var u = B({}, e, {
            "data-href": e.href,
            "data-precedence": e.precedence,
            href: null,
            precedence: null
          });
          return a = (t.ownerDocument || t).createElement(
            "style"
          ), wt(a), Pt(a, "style", u), tc(a, e.precedence, t), l.instance = a;
        case "stylesheet":
          u = $a(e.href);
          var n = t.querySelector(
            Gu(u)
          );
          if (n)
            return l.state.loading |= 4, l.instance = n, wt(n), n;
          a = cd(e), (u = ql.get(u)) && mf(a, u), n = (t.ownerDocument || t).createElement("link"), wt(n);
          var c = n;
          return c._p = new Promise(function(f, s) {
            c.onload = f, c.onerror = s;
          }), Pt(n, "link", a), l.state.loading |= 4, tc(n, e.precedence, t), l.instance = n;
        case "script":
          return n = Wa(e.src), (u = t.querySelector(
            Xu(n)
          )) ? (l.instance = u, wt(u), u) : (a = e, (u = ql.get(n)) && (a = B({}, e), gf(a, u)), t = t.ownerDocument || t, u = t.createElement("script"), wt(u), Pt(u, "link", a), t.head.appendChild(u), l.instance = u);
        case "void":
          return null;
        default:
          throw Error(r(443, l.type));
      }
    else
      l.type === "stylesheet" && (l.state.loading & 4) === 0 && (a = l.instance, l.state.loading |= 4, tc(a, e.precedence, t));
    return l.instance;
  }
  function tc(t, l, e) {
    for (var a = e.querySelectorAll(
      'link[rel="stylesheet"][data-precedence],style[data-precedence]'
    ), u = a.length ? a[a.length - 1] : null, n = u, c = 0; c < a.length; c++) {
      var f = a[c];
      if (f.dataset.precedence === l) n = f;
      else if (n !== u) break;
    }
    n ? n.parentNode.insertBefore(t, n.nextSibling) : (l = e.nodeType === 9 ? e.head : e, l.insertBefore(t, l.firstChild));
  }
  function mf(t, l) {
    t.crossOrigin == null && (t.crossOrigin = l.crossOrigin), t.referrerPolicy == null && (t.referrerPolicy = l.referrerPolicy), t.title == null && (t.title = l.title);
  }
  function gf(t, l) {
    t.crossOrigin == null && (t.crossOrigin = l.crossOrigin), t.referrerPolicy == null && (t.referrerPolicy = l.referrerPolicy), t.integrity == null && (t.integrity = l.integrity);
  }
  var lc = null;
  function fd(t, l, e) {
    if (lc === null) {
      var a = /* @__PURE__ */ new Map(), u = lc = /* @__PURE__ */ new Map();
      u.set(e, a);
    } else
      u = lc, a = u.get(e), a || (a = /* @__PURE__ */ new Map(), u.set(e, a));
    if (a.has(t)) return a;
    for (a.set(t, null), e = e.getElementsByTagName(t), u = 0; u < e.length; u++) {
      var n = e[u];
      if (!(n[Ia] || n[It] || t === "link" && n.getAttribute("rel") === "stylesheet") && n.namespaceURI !== "http://www.w3.org/2000/svg") {
        var c = n.getAttribute(l) || "";
        c = t + c;
        var f = a.get(c);
        f ? f.push(n) : a.set(c, [n]);
      }
    }
    return a;
  }
  function sd(t, l, e) {
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
  function rd(t) {
    return !(t.type === "stylesheet" && (t.state.loading & 3) === 0);
  }
  var Qu = null;
  function Q0() {
  }
  function Z0(t, l, e) {
    if (Qu === null) throw Error(r(475));
    var a = Qu;
    if (l.type === "stylesheet" && (typeof e.media != "string" || matchMedia(e.media).matches !== !1) && (l.state.loading & 4) === 0) {
      if (l.instance === null) {
        var u = $a(e.href), n = t.querySelector(
          Gu(u)
        );
        if (n) {
          t = n._p, t !== null && typeof t == "object" && typeof t.then == "function" && (a.count++, a = ec.bind(a), t.then(a, a)), l.state.loading |= 4, l.instance = n, wt(n);
          return;
        }
        n = t.ownerDocument || t, e = cd(e), (u = ql.get(u)) && mf(e, u), n = n.createElement("link"), wt(n);
        var c = n;
        c._p = new Promise(function(f, s) {
          c.onload = f, c.onerror = s;
        }), Pt(n, "link", e), l.instance = n;
      }
      a.stylesheets === null && (a.stylesheets = /* @__PURE__ */ new Map()), a.stylesheets.set(l, t), (t = l.state.preload) && (l.state.loading & 3) === 0 && (a.count++, l = ec.bind(a), t.addEventListener("load", l), t.addEventListener("error", l));
    }
  }
  function L0() {
    if (Qu === null) throw Error(r(475));
    var t = Qu;
    return t.stylesheets && t.count === 0 && Sf(t, t.stylesheets), 0 < t.count ? function(l) {
      var e = setTimeout(function() {
        if (t.stylesheets && Sf(t, t.stylesheets), t.unsuspend) {
          var a = t.unsuspend;
          t.unsuspend = null, a();
        }
      }, 6e4);
      return t.unsuspend = l, function() {
        t.unsuspend = null, clearTimeout(e);
      };
    } : null;
  }
  function ec() {
    if (this.count--, this.count === 0) {
      if (this.stylesheets) Sf(this, this.stylesheets);
      else if (this.unsuspend) {
        var t = this.unsuspend;
        this.unsuspend = null, t();
      }
    }
  }
  var ac = null;
  function Sf(t, l) {
    t.stylesheets = null, t.unsuspend !== null && (t.count++, ac = /* @__PURE__ */ new Map(), l.forEach(V0, t), ac = null, ec.call(t));
  }
  function V0(t, l) {
    if (!(l.state.loading & 4)) {
      var e = ac.get(t);
      if (e) var a = e.get(null);
      else {
        e = /* @__PURE__ */ new Map(), ac.set(t, e);
        for (var u = t.querySelectorAll(
          "link[data-precedence],style[data-precedence]"
        ), n = 0; n < u.length; n++) {
          var c = u[n];
          (c.nodeName === "LINK" || c.getAttribute("media") !== "not all") && (e.set(c.dataset.precedence, c), a = c);
        }
        a && e.set(null, a);
      }
      u = l.instance, c = u.getAttribute("data-precedence"), n = e.get(c) || a, n === a && e.set(null, u), e.set(c, u), this.count++, a = ec.bind(this), u.addEventListener("load", a), u.addEventListener("error", a), n ? n.parentNode.insertBefore(u, n.nextSibling) : (t = t.nodeType === 9 ? t.head : t, t.insertBefore(u, t.firstChild)), l.state.loading |= 4;
    }
  }
  var Zu = {
    $$typeof: Et,
    Provider: null,
    Consumer: null,
    _currentValue: V,
    _currentValue2: V,
    _threadCount: 0
  };
  function w0(t, l, e, a, u, n, c, f) {
    this.tag = 1, this.containerInfo = t, this.pingCache = this.current = this.pendingChildren = null, this.timeoutHandle = -1, this.callbackNode = this.next = this.pendingContext = this.context = this.cancelPendingCommit = null, this.callbackPriority = 0, this.expirationTimes = At(-1), this.entangledLanes = this.shellSuspendCounter = this.errorRecoveryDisabledLanes = this.expiredLanes = this.warmLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0, this.entanglements = At(0), this.hiddenUpdates = At(null), this.identifierPrefix = a, this.onUncaughtError = u, this.onCaughtError = n, this.onRecoverableError = c, this.pooledCache = null, this.pooledCacheLanes = 0, this.formState = f, this.incompleteTransitions = /* @__PURE__ */ new Map();
  }
  function od(t, l, e, a, u, n, c, f, s, m, j, A) {
    return t = new w0(
      t,
      l,
      e,
      c,
      f,
      s,
      m,
      A
    ), l = 1, n === !0 && (l |= 24), n = Sl(3, null, null, l), t.current = n, n.stateNode = t, l = Pc(), l.refCount++, t.pooledCache = l, l.refCount++, n.memoizedState = {
      element: a,
      isDehydrated: e,
      cache: l
    }, ei(n), t;
  }
  function dd(t) {
    return t ? (t = Ma, t) : Ma;
  }
  function hd(t, l, e, a, u, n) {
    u = dd(u), a.context === null ? a.context = u : a.pendingContext = u, a = Te(l), a.payload = { element: e }, n = n === void 0 ? null : n, n !== null && (a.callback = n), e = Ae(t, a, l), e !== null && (Tl(e, t, l), Su(e, t, l));
  }
  function yd(t, l) {
    if (t = t.memoizedState, t !== null && t.dehydrated !== null) {
      var e = t.retryLane;
      t.retryLane = e !== 0 && e < l ? e : l;
    }
  }
  function bf(t, l) {
    yd(t, l), (t = t.alternate) && yd(t, l);
  }
  function vd(t) {
    if (t.tag === 13) {
      var l = za(t, 67108864);
      l !== null && Tl(l, t, 67108864), bf(t, 67108864);
    }
  }
  var uc = !0;
  function K0(t, l, e, a) {
    var u = p.T;
    p.T = null;
    var n = C.p;
    try {
      C.p = 2, xf(t, l, e, a);
    } finally {
      C.p = n, p.T = u;
    }
  }
  function J0(t, l, e, a) {
    var u = p.T;
    p.T = null;
    var n = C.p;
    try {
      C.p = 8, xf(t, l, e, a);
    } finally {
      C.p = n, p.T = u;
    }
  }
  function xf(t, l, e, a) {
    if (uc) {
      var u = pf(a);
      if (u === null)
        cf(
          t,
          l,
          a,
          nc,
          e
        ), gd(t, a);
      else if ($0(
        u,
        t,
        l,
        e,
        a
      ))
        a.stopPropagation();
      else if (gd(t, a), l & 4 && -1 < k0.indexOf(t)) {
        for (; u !== null; ) {
          var n = ya(u);
          if (n !== null)
            switch (n.tag) {
              case 3:
                if (n = n.stateNode, n.current.memoizedState.isDehydrated) {
                  var c = Ml(n.pendingLanes);
                  if (c !== 0) {
                    var f = n;
                    for (f.pendingLanes |= 2, f.entangledLanes |= 2; c; ) {
                      var s = 1 << 31 - ul(c);
                      f.entanglements[1] |= s, c &= ~s;
                    }
                    Pl(n), (ht & 6) === 0 && (Zn = al() + 500, Hu(0));
                  }
                }
                break;
              case 13:
                f = za(n, 2), f !== null && Tl(f, n, 2), Vn(), bf(n, 2);
            }
          if (n = pf(a), n === null && cf(
            t,
            l,
            a,
            nc,
            e
          ), n === u) break;
          u = n;
        }
        u !== null && a.stopPropagation();
      } else
        cf(
          t,
          l,
          a,
          null,
          e
        );
    }
  }
  function pf(t) {
    return t = Ec(t), jf(t);
  }
  var nc = null;
  function jf(t) {
    if (nc = null, t = ha(t), t !== null) {
      var l = D(t);
      if (l === null) t = null;
      else {
        var e = l.tag;
        if (e === 13) {
          if (t = _(l), t !== null) return t;
          t = null;
        } else if (e === 3) {
          if (l.stateNode.current.memoizedState.isDehydrated)
            return l.tag === 3 ? l.stateNode.containerInfo : null;
          t = null;
        } else l !== t && (t = null);
      }
    }
    return nc = t, null;
  }
  function md(t) {
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
        switch (Se()) {
          case Zl:
            return 2;
          case cl:
            return 8;
          case Ll:
          case Vl:
            return 32;
          case Rt:
            return 268435456;
          default:
            return 32;
        }
      default:
        return 32;
    }
  }
  var Tf = !1, Be = null, Ye = null, Ge = null, Lu = /* @__PURE__ */ new Map(), Vu = /* @__PURE__ */ new Map(), Xe = [], k0 = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(
    " "
  );
  function gd(t, l) {
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
        Lu.delete(l.pointerId);
        break;
      case "gotpointercapture":
      case "lostpointercapture":
        Vu.delete(l.pointerId);
    }
  }
  function wu(t, l, e, a, u, n) {
    return t === null || t.nativeEvent !== n ? (t = {
      blockedOn: l,
      domEventName: e,
      eventSystemFlags: a,
      nativeEvent: n,
      targetContainers: [u]
    }, l !== null && (l = ya(l), l !== null && vd(l)), t) : (t.eventSystemFlags |= a, l = t.targetContainers, u !== null && l.indexOf(u) === -1 && l.push(u), t);
  }
  function $0(t, l, e, a, u) {
    switch (l) {
      case "focusin":
        return Be = wu(
          Be,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "dragenter":
        return Ye = wu(
          Ye,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "mouseover":
        return Ge = wu(
          Ge,
          t,
          l,
          e,
          a,
          u
        ), !0;
      case "pointerover":
        var n = u.pointerId;
        return Lu.set(
          n,
          wu(
            Lu.get(n) || null,
            t,
            l,
            e,
            a,
            u
          )
        ), !0;
      case "gotpointercapture":
        return n = u.pointerId, Vu.set(
          n,
          wu(
            Vu.get(n) || null,
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
  function Sd(t) {
    var l = ha(t.target);
    if (l !== null) {
      var e = D(l);
      if (e !== null) {
        if (l = e.tag, l === 13) {
          if (l = _(e), l !== null) {
            t.blockedOn = l, Zd(t.priority, function() {
              if (e.tag === 13) {
                var a = jl();
                a = Jl(a);
                var u = za(e, a);
                u !== null && Tl(u, e, a), bf(e, a);
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
  function cc(t) {
    if (t.blockedOn !== null) return !1;
    for (var l = t.targetContainers; 0 < l.length; ) {
      var e = pf(t.nativeEvent);
      if (e === null) {
        e = t.nativeEvent;
        var a = new e.constructor(
          e.type,
          e
        );
        Ac = a, e.target.dispatchEvent(a), Ac = null;
      } else
        return l = ya(e), l !== null && vd(l), t.blockedOn = e, !1;
      l.shift();
    }
    return !0;
  }
  function bd(t, l, e) {
    cc(t) && e.delete(l);
  }
  function W0() {
    Tf = !1, Be !== null && cc(Be) && (Be = null), Ye !== null && cc(Ye) && (Ye = null), Ge !== null && cc(Ge) && (Ge = null), Lu.forEach(bd), Vu.forEach(bd);
  }
  function ic(t, l) {
    t.blockedOn === l && (t.blockedOn = null, Tf || (Tf = !0, o.unstable_scheduleCallback(
      o.unstable_NormalPriority,
      W0
    )));
  }
  var fc = null;
  function xd(t) {
    fc !== t && (fc = t, o.unstable_scheduleCallback(
      o.unstable_NormalPriority,
      function() {
        fc === t && (fc = null);
        for (var l = 0; l < t.length; l += 3) {
          var e = t[l], a = t[l + 1], u = t[l + 2];
          if (typeof a != "function") {
            if (jf(a || e) === null)
              continue;
            break;
          }
          var n = ya(e);
          n !== null && (t.splice(l, 3), l -= 3, pi(
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
  function Ku(t) {
    function l(s) {
      return ic(s, t);
    }
    Be !== null && ic(Be, t), Ye !== null && ic(Ye, t), Ge !== null && ic(Ge, t), Lu.forEach(l), Vu.forEach(l);
    for (var e = 0; e < Xe.length; e++) {
      var a = Xe[e];
      a.blockedOn === t && (a.blockedOn = null);
    }
    for (; 0 < Xe.length && (e = Xe[0], e.blockedOn === null); )
      Sd(e), e.blockedOn === null && Xe.shift();
    if (e = (t.ownerDocument || t).$$reactFormReplay, e != null)
      for (a = 0; a < e.length; a += 3) {
        var u = e[a], n = e[a + 1], c = u[sl] || null;
        if (typeof n == "function")
          c || xd(e);
        else if (c) {
          var f = null;
          if (n && n.hasAttribute("formAction")) {
            if (u = n, c = n[sl] || null)
              f = c.formAction;
            else if (jf(u) !== null) continue;
          } else f = c.action;
          typeof f == "function" ? e[a + 1] = f : (e.splice(a, 3), a -= 3), xd(e);
        }
      }
  }
  function Af(t) {
    this._internalRoot = t;
  }
  sc.prototype.render = Af.prototype.render = function(t) {
    var l = this._internalRoot;
    if (l === null) throw Error(r(409));
    var e = l.current, a = jl();
    hd(e, a, t, l, null, null);
  }, sc.prototype.unmount = Af.prototype.unmount = function() {
    var t = this._internalRoot;
    if (t !== null) {
      this._internalRoot = null;
      var l = t.containerInfo;
      hd(t.current, 2, null, t, null, null), Vn(), l[da] = null;
    }
  };
  function sc(t) {
    this._internalRoot = t;
  }
  sc.prototype.unstable_scheduleHydration = function(t) {
    if (t) {
      var l = Iu();
      t = { blockedOn: null, target: t, priority: l };
      for (var e = 0; e < Xe.length && l !== 0 && l < Xe[e].priority; e++) ;
      Xe.splice(e, 0, t), e === 0 && Sd(t);
    }
  };
  var pd = b.version;
  if (pd !== "19.1.1")
    throw Error(
      r(
        527,
        pd,
        "19.1.1"
      )
    );
  C.findDOMNode = function(t) {
    var l = t._reactInternals;
    if (l === void 0)
      throw typeof t.render == "function" ? Error(r(188)) : (t = Object.keys(t).join(","), Error(r(268, t)));
    return t = N(l), t = t !== null ? x(t) : null, t = t === null ? null : t.stateNode, t;
  };
  var F0 = {
    bundleType: 0,
    version: "19.1.1",
    rendererPackageName: "react-dom",
    currentDispatcherRef: p,
    reconcilerVersion: "19.1.1"
  };
  if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ < "u") {
    var rc = __REACT_DEVTOOLS_GLOBAL_HOOK__;
    if (!rc.isDisabled && rc.supportsFiber)
      try {
        zl = rc.inject(
          F0
        ), nt = rc;
      } catch {
      }
  }
  return ku.createRoot = function(t, l) {
    if (!U(t)) throw Error(r(299));
    var e = !1, a = "", u = qr, n = Br, c = Yr, f = null;
    return l != null && (l.unstable_strictMode === !0 && (e = !0), l.identifierPrefix !== void 0 && (a = l.identifierPrefix), l.onUncaughtError !== void 0 && (u = l.onUncaughtError), l.onCaughtError !== void 0 && (n = l.onCaughtError), l.onRecoverableError !== void 0 && (c = l.onRecoverableError), l.unstable_transitionCallbacks !== void 0 && (f = l.unstable_transitionCallbacks)), l = od(
      t,
      1,
      !1,
      null,
      null,
      e,
      a,
      u,
      n,
      c,
      f,
      null
    ), t[da] = l.current, nf(t), new Af(l);
  }, ku.hydrateRoot = function(t, l, e) {
    if (!U(t)) throw Error(r(299));
    var a = !1, u = "", n = qr, c = Br, f = Yr, s = null, m = null;
    return e != null && (e.unstable_strictMode === !0 && (a = !0), e.identifierPrefix !== void 0 && (u = e.identifierPrefix), e.onUncaughtError !== void 0 && (n = e.onUncaughtError), e.onCaughtError !== void 0 && (c = e.onCaughtError), e.onRecoverableError !== void 0 && (f = e.onRecoverableError), e.unstable_transitionCallbacks !== void 0 && (s = e.unstable_transitionCallbacks), e.formState !== void 0 && (m = e.formState)), l = od(
      t,
      1,
      !0,
      l,
      e ?? null,
      a,
      u,
      n,
      c,
      f,
      s,
      m
    ), l.context = dd(null), e = l.current, a = jl(), a = Jl(a), u = Te(a), u.callback = null, Ae(e, u, a), e = a, l.current.lanes = e, Ct(l, e), Pl(l), t[da] = l.current, nf(t), new sc(l);
  }, ku.version = "19.1.1", ku;
}
var Rd;
function i1() {
  if (Rd) return zf.exports;
  Rd = 1;
  function o() {
    if (!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ > "u" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE != "function"))
      try {
        __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o);
      } catch (b) {
        console.error(b);
      }
  }
  return o(), zf.exports = c1(), zf.exports;
}
var f1 = i1(), R = _f();
const s1 = (o) => o.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase(), qd = (...o) => o.filter((b, z, r) => !!b && b.trim() !== "" && r.indexOf(b) === z).join(" ").trim();
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
    strokeWidth: z = 2,
    absoluteStrokeWidth: r,
    className: U = "",
    children: D,
    iconNode: _,
    ...O
  }, N) => R.createElement(
    "svg",
    {
      ref: N,
      ...r1,
      width: b,
      height: b,
      stroke: o,
      strokeWidth: r ? Number(z) * 24 / Number(b) : z,
      className: qd("lucide", U),
      ...O
    },
    [
      ..._.map(([x, B]) => R.createElement(x, B)),
      ...Array.isArray(D) ? D : [D]
    ]
  )
);
const bt = (o, b) => {
  const z = R.forwardRef(
    ({ className: r, ...U }, D) => R.createElement(o1, {
      ref: D,
      iconNode: b,
      className: qd(`lucide-${s1(o)}`, r),
      ...U
    })
  );
  return z.displayName = `${o}`, z;
};
const Rf = bt("AudioLines", [
  ["path", { d: "M2 10v3", key: "1fnikh" }],
  ["path", { d: "M6 6v11", key: "11sgs0" }],
  ["path", { d: "M10 3v18", key: "yhl04a" }],
  ["path", { d: "M14 8v7", key: "3a1oy3" }],
  ["path", { d: "M18 5v13", key: "123xd1" }],
  ["path", { d: "M22 10v3", key: "154ddg" }]
]);
const $u = bt("Check", [["path", { d: "M20 6 9 17l-5-5", key: "1gmf2c" }]]);
const hc = bt("ChevronDown", [
  ["path", { d: "m6 9 6 6 6-6", key: "qrunsl" }]
]);
const Uf = bt("ChevronRight", [
  ["path", { d: "m9 18 6-6-6-6", key: "mthhwq" }]
]);
const d1 = bt("ChevronUp", [["path", { d: "m18 15-6-6-6 6", key: "153udz" }]]);
const Bd = bt("CircleCheck", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["path", { d: "m9 12 2 2 4-4", key: "dzmm74" }]
]);
const h1 = bt("Circle", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }]
]);
const y1 = bt("Clock3", [
  ["circle", { cx: "12", cy: "12", r: "10", key: "1mglay" }],
  ["polyline", { points: "12 6 12 12 16.5 12", key: "1aq6pp" }]
]);
const v1 = bt("CodeXml", [
  ["path", { d: "m18 16 4-4-4-4", key: "1inbqp" }],
  ["path", { d: "m6 8-4 4 4 4", key: "15zrgr" }],
  ["path", { d: "m14.5 4-5 16", key: "e7oirm" }]
]);
const m1 = bt("Ellipsis", [
  ["circle", { cx: "12", cy: "12", r: "1", key: "41hilf" }],
  ["circle", { cx: "19", cy: "12", r: "1", key: "1wjl8i" }],
  ["circle", { cx: "5", cy: "12", r: "1", key: "1pcz8c" }]
]);
const Cf = bt("FileAudio", [
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
const Hf = bt("FolderGit2", [
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
const me = bt("LoaderCircle", [
  ["path", { d: "M21 12a9 9 0 1 1-6.219-8.56", key: "13zald" }]
]);
const g1 = bt("Mic", [
  ["path", { d: "M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z", key: "131961" }],
  ["path", { d: "M19 10v2a7 7 0 0 1-14 0v-2", key: "1vc78b" }],
  ["line", { x1: "12", x2: "12", y1: "19", y2: "22", key: "x3vr5v" }]
]);
const S1 = bt("Pause", [
  ["rect", { x: "14", y: "4", width: "4", height: "16", rx: "1", key: "zuxfzm" }],
  ["rect", { x: "6", y: "4", width: "4", height: "16", rx: "1", key: "1okwgv" }]
]);
const oc = bt("PencilLine", [
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
const b1 = bt("Play", [
  ["polygon", { points: "6 3 20 12 6 21 6 3", key: "1oa8hb" }]
]);
const dc = bt("Plus", [
  ["path", { d: "M5 12h14", key: "1ays0h" }],
  ["path", { d: "M12 5v14", key: "s699le" }]
]);
const Yd = bt("Radio", [
  ["path", { d: "M4.9 19.1C1 15.2 1 8.8 4.9 4.9", key: "1vaf9d" }],
  ["path", { d: "M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5", key: "u1ii0m" }],
  ["circle", { cx: "12", cy: "12", r: "2", key: "1c9p78" }],
  ["path", { d: "M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5", key: "1j5fej" }],
  ["path", { d: "M19.1 4.9C23 8.8 23 15.1 19.1 19", key: "10b0cb" }]
]);
const x1 = bt("RefreshCw", [
  ["path", { d: "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8", key: "v9h5vc" }],
  ["path", { d: "M21 3v5h-5", key: "1q7to0" }],
  ["path", { d: "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16", key: "3uifl3" }],
  ["path", { d: "M8 16H3v5", key: "1cv678" }]
]);
const p1 = bt("Search", [
  ["circle", { cx: "11", cy: "11", r: "8", key: "4ej97u" }],
  ["path", { d: "m21 21-4.3-4.3", key: "1qie3q" }]
]);
const j1 = bt("Send", [
  [
    "path",
    {
      d: "M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z",
      key: "1ffxy3"
    }
  ],
  ["path", { d: "m21.854 2.147-10.94 10.939", key: "12cjpa" }]
]);
const T1 = bt("Settings2", [
  ["path", { d: "M20 7h-9", key: "3s1dr2" }],
  ["path", { d: "M14 17H5", key: "gfn3mx" }],
  ["circle", { cx: "17", cy: "17", r: "3", key: "18b49y" }],
  ["circle", { cx: "7", cy: "7", r: "3", key: "dfmy0x" }]
]);
const _d = bt("Sparkles", [
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
const A1 = bt("Trash2", [
  ["path", { d: "M3 6h18", key: "d0wm0j" }],
  ["path", { d: "M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6", key: "4alrt4" }],
  ["path", { d: "M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2", key: "v07s0e" }],
  ["line", { x1: "10", x2: "10", y1: "11", y2: "17", key: "1uufr5" }],
  ["line", { x1: "14", x2: "14", y1: "11", y2: "17", key: "xtxkd" }]
]);
const E1 = bt("X", [
  ["path", { d: "M18 6 6 18", key: "1bl5f8" }],
  ["path", { d: "m6 6 12 12", key: "d8bk6v" }]
]);
const z1 = bt("Zap", [
  [
    "path",
    {
      d: "M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z",
      key: "1xq2db"
    }
  ]
]);
let yc = async () => {
  throw Error("请从已连接的设备打开 SaySo");
};
function M1(o) {
  yc = o;
}
function N1(o, b, z) {
  return yc(o, b, z);
}
function D1(o) {
  return Uint8Array.from(atob(o), (b) => b.charCodeAt(0));
}
let Ud = Promise.resolve();
function O1(o, b, z) {
  const r = yc, U = Ud.catch(() => {
  }).then(async () => {
    const { id: D } = await r("POST", "/api/audio/begin", {
      ...z,
      sessionId: o,
      sizeBytes: b.size,
      mimeType: b.type || "audio/webm"
    });
    try {
      for (let _ = 0; _ < b.size; _ += 16e3) {
        const O = new Uint8Array(
          await b.slice(_, _ + 16e3).arrayBuffer()
        );
        await r("POST", `/api/audio/${D}/chunk`, {
          offset: _,
          data: btoa(String.fromCharCode(...O))
        });
      }
      return await r("POST", `/api/audio/${D}/finish`);
    } catch (_) {
      throw await r("POST", `/api/audio/${D}/abort`).catch(
        () => {
        }
      ), _;
    }
  });
  return Ud = U, U;
}
async function R1(o, b) {
  const z = yc, r = [];
  let U = 0, D = 0, _ = "";
  do {
    if (b.aborted) throw Error("已取消");
    const O = await z("POST", `/api/audio/${o}/read`, {
      offset: U
    });
    if (!O.size || O.size > 64 * 1024 * 1024)
      throw Error("录音大小不合法");
    const N = D1(O.data);
    if (!N.length) throw Error("录音数据不完整");
    r.push(N), U += N.length, D = O.size, _ = O.mimeType;
  } while (U < D);
  if (b.aborted) throw Error("已取消");
  return URL.createObjectURL(new Blob(r, { type: _ }));
}
async function yl(o, b) {
  return N1(
    b?.method || "GET",
    o,
    b?.body ? JSON.parse(String(b.body)) : void 0
  );
}
const Zt = {
  cancelRun: (o) => yl(`/api/runs/${o}/cancel`, { method: "POST" }),
  state: () => yl("/api/state"),
  updateSettings: (o) => yl("/api/settings", {
    method: "PATCH",
    body: JSON.stringify(o)
  }),
  createProject: (o) => yl("/api/projects", {
    method: "POST",
    body: JSON.stringify(o)
  }),
  updateProject: (o, b) => yl(`/api/projects/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  createSession: (o, b) => yl("/api/sessions", {
    method: "POST",
    body: JSON.stringify({ projectId: o, title: b })
  }),
  updateSession: (o, b) => yl(`/api/sessions/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  deleteSession: (o) => yl(`/api/sessions/${o}`, { method: "DELETE" }),
  addUtterance: (o, b) => yl(`/api/sessions/${o}/utterances`, {
    method: "POST",
    body: JSON.stringify(b)
  }),
  extractActions: (o) => yl(`/api/sessions/${o}/extract-actions`, { method: "POST" }),
  createAction: (o, b) => yl(`/api/sessions/${o}/actions`, {
    method: "POST",
    body: JSON.stringify(b)
  }),
  updateAction: (o, b) => yl(`/api/actions/${o}`, {
    method: "PATCH",
    body: JSON.stringify(b)
  }),
  deleteAction: (o) => yl(`/api/actions/${o}`, { method: "DELETE" }),
  dispatchAction: (o) => yl(`/api/actions/${o}/dispatch`, { method: "POST" }),
  uploadAudio: O1
};
function qf(o) {
  return o.trim().toLocaleLowerCase();
}
function _1(o, b) {
  const z = qf(b);
  return z.length > 0 && o.toLocaleLowerCase().includes(z);
}
function U1(o, b) {
  const z = qf(b);
  if (!z) return [{ text: o, match: !1 }];
  const r = o.toLocaleLowerCase(), U = [];
  let D = 0;
  for (; D < o.length; ) {
    const _ = r.indexOf(z, D);
    if (_ === -1) {
      U.push({ text: o.slice(D), match: !1 });
      break;
    }
    _ > D && U.push({ text: o.slice(D, _), match: !1 }), U.push({
      text: o.slice(_, _ + z.length),
      match: !0
    }), D = _ + z.length;
  }
  return U.length ? U : [{ text: o, match: !1 }];
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
function Gd(o) {
  const b = Math.max(0, Math.floor(o / 1e3));
  return `${String(Math.floor(b / 60)).padStart(2, "0")}:${String(b % 60).padStart(2, "0")}`;
}
function Cd(o) {
  const b = new Date(o);
  return Number.isNaN(b.getTime()) ? o : `${b.getFullYear()}-${b.getMonth()}-${b.getDate()}-${b.getHours()}-${b.getMinutes()}`;
}
function G1(o) {
  const b = new Date(o);
  return Number.isNaN(b.getTime()) ? "--:--" : `${String(b.getHours()).padStart(2, "0")}:${String(b.getMinutes()).padStart(2, "0")}`;
}
function X1(o) {
  return o ? o < 1024 * 1024 ? `${Math.max(1, Math.round(o / 1024))} KB` : `${(o / 1024 / 1024).toFixed(1)} MB` : "大小未知";
}
function Hd(o) {
  const b = new Date(o), z = /* @__PURE__ */ new Date();
  if (b.toDateString() === z.toDateString()) return "今天";
  const r = new Date(z);
  return r.setDate(z.getDate() - 1), b.toDateString() === r.toDateString() ? "昨天" : `${b.getMonth() + 1}月${b.getDate()}日`;
}
function ia({
  title: o,
  description: b,
  children: z,
  onClose: r,
  wide: U = !1
}) {
  return /* @__PURE__ */ i.jsx(
    "div",
    {
      className: "modal-backdrop",
      onMouseDown: (D) => D.target === D.currentTarget && r(),
      children: /* @__PURE__ */ i.jsxs(
        "section",
        {
          className: `modal ${U ? "modal-wide" : ""}`,
          role: "dialog",
          "aria-modal": "true",
          children: [
            /* @__PURE__ */ i.jsxs("div", { className: "modal-heading", children: [
              /* @__PURE__ */ i.jsxs("div", { children: [
                /* @__PURE__ */ i.jsx("h2", { children: o }),
                b && /* @__PURE__ */ i.jsx("p", { children: b })
              ] }),
              /* @__PURE__ */ i.jsx("button", { className: "icon-button", onClick: r, "aria-label": "关闭", children: /* @__PURE__ */ i.jsx(E1, { size: 18 }) })
            ] }),
            z
          ]
        }
      )
    }
  );
}
function Xd({ status: o }) {
  const b = {
    draft: "待完善",
    ready: "Ready",
    running: "Codex 工作中",
    done: "已完成",
    failed: "执行失败"
  }[o];
  return /* @__PURE__ */ i.jsxs("span", { className: `status-pill status-${o}`, children: [
    o === "running" && /* @__PURE__ */ i.jsx(me, { size: 12 }),
    " ",
    b
  ] });
}
function Qd(o, b) {
  const [z, r] = R.useState([]), U = o.join("\0");
  return R.useEffect(() => {
    if (!b) {
      r([]);
      return;
    }
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      r(o);
      return;
    }
    let D = !1, _ = 0, O = 0, N;
    r([]);
    const x = () => {
      if (D) return;
      const B = Array.from(o[_]);
      if (O < B.length) {
        O += 1;
        const X = B.slice(0, O).join("");
        r((W) => {
          const it = W.slice();
          return it[_] = X, it;
        }), N = window.setTimeout(x, 34);
        return;
      }
      if (_ < o.length - 1) {
        _ += 1, O = 0, N = window.setTimeout(x, 360);
        return;
      }
      N = window.setTimeout(() => {
        _ = 0, O = 0, r([]), N = window.setTimeout(x, 240);
      }, 1600);
    };
    return N = window.setTimeout(x, 160), () => {
      D = !0, N !== void 0 && window.clearTimeout(N);
    };
  }, [b, U]), z;
}
function Q1({
  onClose: o,
  deviceLabel: b
}) {
  const [z, r] = R.useState(!1), [U, D] = R.useState([]), _ = R.useRef([]), O = R.useRef(/* @__PURE__ */ new Set()), N = R.useRef(async () => {
  }), [x, B] = R.useState(""), [X, W] = R.useState(!1), [it, q] = R.useState(
    "discussion"
  ), [ot, Dt] = R.useState(!1);
  R.useEffect(() => {
    const E = (F) => {
      (ml.current || O.current.size) && (F.preventDefault(), F.returnValue = "");
    }, Q = () => {
      N.current().catch(() => {
      }).finally(() => Rt("设备连接已断开，未上传的录音可下载保存。"));
    };
    return window.addEventListener("beforeunload", E), window.addEventListener("remote-agent-disconnected", Q), () => {
      Se.current = !0;
      try {
        al.current?.stop();
      } catch {
      }
      Zl.current?.getTracks().forEach((F) => F.stop()), cl.current && window.clearTimeout(cl.current), _.current.forEach((F) => URL.revokeObjectURL(F.url)), window.removeEventListener("beforeunload", E), window.removeEventListener("remote-agent-disconnected", Q);
    };
  }, []);
  async function xt() {
    r(!0);
    try {
      if (await N.current(), await Promise.allSettled([...O.current]), _.current.length && !confirm("仍有未上传录音，请先下载保存。确定离开吗？"))
        return;
      o();
    } catch (E) {
      Rt(
        E instanceof Error ? E.message : "正在保存，请稍后再试"
      );
    } finally {
      r(!1);
    }
  }
  const [ft, Et] = R.useState(q1), [Lt, P] = R.useState(!0), [Vt, Z] = R.useState(""), [zt, $t] = R.useState(""), [ge, Al] = R.useState(
    /* @__PURE__ */ new Set()
  ), [Ot, vl] = R.useState(
    null
  ), [Il, Bt] = R.useState(!1), [p, C] = R.useState(null), [V, st] = R.useState(
    null
  ), [d, M] = R.useState(null), [Y, H] = R.useState(null), [w, ut] = R.useState(""), [$, el] = R.useState(""), [pt, El] = R.useState(""), [Bl, Ze] = R.useState(!1), [te, fa] = R.useState(""), [Fa, Wu] = R.useState(0), ml = R.useRef(null), al = R.useRef(
    null
  ), Se = R.useRef(!1), Zl = R.useRef(null), cl = R.useRef(null), Ll = R.useRef(null), Vl = R.useRef(null);
  function Rt(E) {
    ut(E), window.setTimeout(() => ut(""), 3600);
  }
  const il = async (E = !1) => {
    try {
      const Q = await Zt.state();
      Et(Q), Z(
        (F) => F && Q.projects.some((jt) => jt.id === F) ? F : Q.projects[0]?.id || ""
      ), $t(
        (F) => F && Q.sessions.some((jt) => jt.id === F) ? F : Q.sessions.find(
          (jt) => jt.projectId === (Vt || Q.projects[0]?.id)
        )?.id || ""
      );
    } catch (Q) {
      E || Rt(Q instanceof Error ? Q.message : "无法连接服务");
    } finally {
      P(!1);
    }
  };
  R.useEffect(() => {
    il();
  }, []);
  const Pa = ft.actions.some((E) => E.status === "running") || ft.sessions.some(
    (E) => E.status === "recording" || ["queued", "running"].includes(E.analysisStatus || "")
  );
  R.useEffect(() => (Pa && !Vl.current && (Vl.current = window.setInterval(() => {
    il(!0);
  }, 1800)), !Pa && Vl.current && (window.clearInterval(Vl.current), Vl.current = null), () => {
    Vl.current && window.clearInterval(Vl.current), Vl.current = null;
  }), [Pa]);
  const zl = ft.projects.find((E) => E.id === Vt) || ft.projects[0], nt = ft.sessions.find(
    (E) => E.id === zt
  ), Yl = ft.actions.filter((E) => E.sessionId === nt?.id).sort((E, Q) => {
    const F = { running: 0, ready: 1, draft: 2, failed: 3, done: 4 };
    return F[E.status] - F[Q.status] || Q.createdAt.localeCompare(E.createdAt);
  }), ul = R.useMemo(
    () => new Map(
      ft.projects.map((E) => [
        E.id,
        ft.sessions.filter((Q) => Q.projectId === E.id)
      ])
    ),
    [ft.projects, ft.sessions]
  );
  R.useEffect(() => {
    if (!Bl || !Ll.current) return;
    const E = () => Wu(
      Date.now() - new Date(Ll.current.startedAt).getTime()
    );
    E();
    const Q = window.setInterval(E, 1e3);
    return () => window.clearInterval(Q);
  }, [Bl]);
  function vc(E) {
    Dt(!1), Z(E), $t(
      ft.sessions.find((Q) => Q.projectId === E)?.id || ""
    ), Al((Q) => {
      const F = new Set(Q);
      return F.delete(E), F;
    });
  }
  async function Fu() {
    if (zl)
      try {
        Bl && await Ml();
        const E = await Zt.createSession(zl.id, "");
        $t(E.id), Dt(!1), await il(!0);
      } catch (E) {
        Rt(E instanceof Error ? E.message : "创建会话失败");
      }
  }
  function Pu(E, Q) {
    const F = Date.now(), jt = [], At = new MediaRecorder(E, { audioBitsPerSecond: 64e3 });
    ml.current = At, At.ondataavailable = (fl) => {
      fl.data.size && jt.push(fl.data);
    };
    let Ct = () => {
    };
    const oa = new Promise((fl) => {
      Ct = fl;
    });
    At.saved = oa, At.onstop = () => {
      const fl = Date.now(), Kl = new Blob(jt, {
        type: At.mimeType || "audio/webm"
      });
      if (!Kl.size) {
        Ct();
        return;
      }
      const Jl = Zt.uploadAudio(Q.id, Kl, {
        startedAt: new Date(F).toISOString(),
        endedAt: new Date(fl).toISOString(),
        durationMs: fl - F
      }).then(() => il(!0)).catch((Le) => {
        const Iu = URL.createObjectURL(Kl);
        _.current.push({
          url: Iu,
          name: `sayso-${new Date(F).toISOString().replace(/:/g, "-")}.${Kl.type.includes("mp4") ? "m4a" : "webm"}`
        }), D([..._.current]), Rt(`录音未上传：${Le.message}。请下载备份。`);
      }).finally(() => {
        O.current.delete(Jl), Ct();
      });
      O.current.add(Jl);
    }, At.start(1e3), cl.current = window.setTimeout(
      () => sa(Q),
      C1
    );
  }
  function sa(E) {
    cl.current && window.clearTimeout(cl.current), cl.current = null;
    const Q = Zl.current, F = ml.current;
    !Q || !F || F.state !== "recording" || (F.stop(), Pu(Q, E));
  }
  async function ra(E) {
    if (Bl && await Ml(), !navigator.mediaDevices?.getUserMedia) {
      Rt("当前浏览器不支持录音");
      return;
    }
    try {
      const Q = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: !0, noiseSuppression: !0 }
      });
      Zl.current = Q, await Zt.updateSession(E.id, { status: "recording" }), Ll.current = E, Se.current = !1, Pu(Q, E);
      const F = window, jt = F.SpeechRecognition || F.webkitSpeechRecognition;
      if (jt) {
        const At = new jt();
        At.continuous = !0, At.interimResults = !0, At.lang = "zh-CN", At.onresult = (Ct) => {
          let oa = "";
          for (let fl = Ct.resultIndex; fl < Ct.results.length; fl += 1) {
            const Kl = Ct.results[fl], Jl = Kl[0].transcript.trim();
            Kl.isFinal && Jl ? Zt.addUtterance(E.id, {
              text: Jl,
              atMs: Date.now() - new Date(E.startedAt).getTime()
            }).then(() => il(!0)).catch((Le) => Rt(Le.message)) : oa += Jl;
          }
          El(oa);
        }, At.onerror = (Ct) => {
          [
            "network",
            "service-not-allowed",
            "not-allowed",
            "audio-capture"
          ].includes(Ct.error) && (Se.current = !0, El(""), Rt(
            Ct.error === "network" ? "实时转写服务网络不可用；录音仍会继续并正常保存" : "录音正常；当前浏览器暂时无法使用实时转写"
          ));
        }, At.onend = () => {
          if (!Se.current && ml.current?.state === "recording")
            try {
              At.start();
            } catch {
            }
        };
        try {
          At.start();
        } catch {
          Se.current = !0, Rt("录音已开始，实时转写不可用，可以打字补充。");
        }
        al.current = At;
      } else Rt("录音已开始；当前浏览器不提供实时转写");
      fa(E.id), Ze(!0), Wu(Date.now() - new Date(E.startedAt).getTime()), await il(!0);
    } catch (Q) {
      cl.current && window.clearTimeout(cl.current), ml.current?.state === "recording" && ml.current.stop(), await Zt.updateSession(E.id, { status: "paused" }).catch(() => {
      }), Zl.current?.getTracks().forEach((F) => F.stop()), Zl.current = null, ml.current = null, Rt(
        Q instanceof Error ? `无法开始录音：${Q.message}` : "无法开始录音"
      );
    }
  }
  async function Ml() {
    const E = Ll.current?.id || te, Q = ml.current?.saved;
    Se.current = !0;
    try {
      al.current?.stop();
    } catch {
    }
    al.current = null, cl.current && window.clearTimeout(cl.current), cl.current = null, ml.current?.state !== "inactive" && ml.current?.stop(), Zl.current?.getTracks().forEach((F) => F.stop()), Zl.current = null, ml.current = null, Ll.current = null, El(""), Ze(!1), fa(""), await Q, await Promise.allSettled([...O.current]), E && (await Zt.updateSession(E, { status: "paused" }), await il(!0));
  }
  N.current = Ml;
  async function wl(E, Q) {
    try {
      await E(), Q && Rt(Q), await il(!0);
    } catch (F) {
      Rt(F instanceof Error ? F.message : "操作失败");
    }
  }
  return Lt ? /* @__PURE__ */ i.jsxs("div", { className: "loading-screen", children: [
    /* @__PURE__ */ i.jsx("div", { className: "brand-mark", children: /* @__PURE__ */ i.jsx(Rf, { size: 22 }) }),
    /* @__PURE__ */ i.jsx("p", { children: "正在打开 SaySo…" }),
    /* @__PURE__ */ i.jsx("button", { onClick: o, children: "返回工作台" })
  ] }) : /* @__PURE__ */ i.jsxs(
    "div",
    {
      className: `sayso-layout pane-${it} ${ot ? "tree-open" : ""}`,
      children: [
        /* @__PURE__ */ i.jsxs("header", { className: "remote-header", children: [
          /* @__PURE__ */ i.jsxs("button", { onClick: () => {
            xt();
          }, disabled: z, children: [
            "← ",
            z ? "正在保存…" : "工作台"
          ] }),
          /* @__PURE__ */ i.jsx("strong", { children: "SaySo" }),
          /* @__PURE__ */ i.jsx("span", { children: b }),
          /* @__PURE__ */ i.jsx(
            "button",
            {
              className: "mobile-tree-button",
              onClick: () => Dt(!ot),
              children: "项目 / 会话"
            }
          )
        ] }),
        U.length > 0 && /* @__PURE__ */ i.jsxs("div", { className: "audio-recovery", children: [
          "未上传录音，请下载保存：",
          U.map((E) => /* @__PURE__ */ i.jsx("a", { href: E.url, download: E.name, children: "下载录音备份" }, E.url))
        ] }),
        /* @__PURE__ */ i.jsxs("div", { className: "app-shell", children: [
          /* @__PURE__ */ i.jsxs("aside", { className: "sidebar", children: [
            /* @__PURE__ */ i.jsxs("div", { className: "brand-row", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "brand", children: [
                /* @__PURE__ */ i.jsx("div", { className: "brand-mark", children: /* @__PURE__ */ i.jsx(Rf, { size: 21 }) }),
                /* @__PURE__ */ i.jsx("span", { children: "SaySo" })
              ] }),
              /* @__PURE__ */ i.jsx(
                "button",
                {
                  className: "system-settings",
                  onClick: () => Bt(!0),
                  "aria-label": "系统设置",
                  children: /* @__PURE__ */ i.jsx(T1, { size: 17 })
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs(
              "button",
              {
                className: "new-session-button",
                onClick: () => {
                  Fu();
                },
                disabled: !zl,
                children: [
                  /* @__PURE__ */ i.jsx(dc, { size: 16 }),
                  " 新会话 ",
                  /* @__PURE__ */ i.jsx("span", { children: "⌘ N" })
                ]
              }
            ),
            /* @__PURE__ */ i.jsxs("div", { className: "tree-heading", children: [
              /* @__PURE__ */ i.jsx("span", { children: "项目" }),
              /* @__PURE__ */ i.jsx(
                "button",
                {
                  onClick: () => vl("new"),
                  "aria-label": "添加项目",
                  children: /* @__PURE__ */ i.jsx(dc, { size: 14 })
                }
              )
            ] }),
            /* @__PURE__ */ i.jsx("div", { className: "project-tree", children: ft.projects.map((E) => {
              const Q = ul.get(E.id) || [], F = ge.has(E.id);
              return /* @__PURE__ */ i.jsxs("div", { className: "project-branch", children: [
                /* @__PURE__ */ i.jsxs(
                  "div",
                  {
                    className: `project-node ${zl?.id === E.id ? "selected" : ""}`,
                    children: [
                      /* @__PURE__ */ i.jsxs(
                        "button",
                        {
                          className: "project-select",
                          onClick: () => vc(E.id),
                          children: [
                            /* @__PURE__ */ i.jsx(
                              "span",
                              {
                                className: "tree-chevron",
                                onClick: (jt) => {
                                  jt.stopPropagation(), Al((At) => {
                                    const Ct = new Set(At);
                                    return Ct.has(E.id) ? Ct.delete(E.id) : Ct.add(E.id), Ct;
                                  });
                                },
                                children: F ? /* @__PURE__ */ i.jsx(Uf, { size: 13 }) : /* @__PURE__ */ i.jsx(hc, { size: 13 })
                              }
                            ),
                            /* @__PURE__ */ i.jsx(Hf, { size: 15 }),
                            /* @__PURE__ */ i.jsx("strong", { children: E.name })
                          ]
                        }
                      ),
                      /* @__PURE__ */ i.jsx(
                        "button",
                        {
                          className: "tree-edit",
                          onClick: () => vl(E),
                          "aria-label": `编辑项目 ${E.name}`,
                          children: /* @__PURE__ */ i.jsx(oc, { size: 13 })
                        }
                      )
                    ]
                  }
                ),
                !F && /* @__PURE__ */ i.jsxs("div", { className: "session-tree", children: [
                  Q.map((jt) => {
                    const At = Bl && te === jt.id;
                    return /* @__PURE__ */ i.jsxs(
                      "div",
                      {
                        className: `tree-session ${nt?.id === jt.id ? "active" : ""}`,
                        children: [
                          /* @__PURE__ */ i.jsxs(
                            "button",
                            {
                              className: "session-select",
                              onClick: () => {
                                Z(E.id), $t(jt.id), Dt(!1);
                              },
                              children: [
                                /* @__PURE__ */ i.jsx(
                                  "span",
                                  {
                                    className: `tree-session-icon ${At ? "recording" : ""}`,
                                    children: At ? /* @__PURE__ */ i.jsx(Yd, { size: 12 }) : /* @__PURE__ */ i.jsx(Cf, { size: 12 })
                                  }
                                ),
                                /* @__PURE__ */ i.jsxs("span", { children: [
                                  /* @__PURE__ */ i.jsx("strong", { children: jt.title }),
                                  /* @__PURE__ */ i.jsxs("small", { children: [
                                    Hd(jt.startedAt),
                                    " ·",
                                    " ",
                                    jt.transcript.length,
                                    " 条"
                                  ] })
                                ] })
                              ]
                            }
                          ),
                          /* @__PURE__ */ i.jsx(
                            "button",
                            {
                              className: "tree-edit",
                              onClick: () => C(jt),
                              "aria-label": `重命名 ${jt.title}`,
                              children: /* @__PURE__ */ i.jsx(oc, { size: 12 })
                            }
                          )
                        ]
                      },
                      jt.id
                    );
                  }),
                  Q.length === 0 && /* @__PURE__ */ i.jsx("div", { className: "tree-empty", children: "暂无会话" })
                ] })
              ] }, E.id);
            }) })
          ] }),
          /* @__PURE__ */ i.jsx("main", { className: "workspace", children: nt ? /* @__PURE__ */ i.jsxs(i.Fragment, { children: [
            /* @__PURE__ */ i.jsxs("header", { className: "topbar", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "title-stack", children: [
                /* @__PURE__ */ i.jsxs("div", { className: "eyebrow", children: [
                  /* @__PURE__ */ i.jsx("span", { children: zl?.name }),
                  /* @__PURE__ */ i.jsx("span", { children: "/" }),
                  /* @__PURE__ */ i.jsx("span", { children: Hd(nt.startedAt) })
                ] }),
                /* @__PURE__ */ i.jsxs("div", { className: "session-title-line", children: [
                  /* @__PURE__ */ i.jsx("h1", { children: nt.title }),
                  /* @__PURE__ */ i.jsx(
                    "button",
                    {
                      className: "small-icon",
                      onClick: () => C(nt),
                      "aria-label": "编辑会话名称",
                      children: /* @__PURE__ */ i.jsx(oc, { size: 14 })
                    }
                  )
                ] })
              ] }),
              /* @__PURE__ */ i.jsxs("div", { className: "topbar-actions", children: [
                (nt.recordings?.length || nt.recordingPath) && /* @__PURE__ */ i.jsxs(
                  "button",
                  {
                    className: "ghost-button",
                    onClick: () => H(nt),
                    children: [
                      /* @__PURE__ */ i.jsx(b1, { size: 15 }),
                      " 录音片段"
                    ]
                  }
                ),
                Bl && te === nt.id ? /* @__PURE__ */ i.jsxs("div", { className: "recording-control", children: [
                  /* @__PURE__ */ i.jsxs("span", { className: "recording-time", children: [
                    /* @__PURE__ */ i.jsx("i", {}),
                    " ",
                    Gd(Fa)
                  ] }),
                  /* @__PURE__ */ i.jsxs(
                    "button",
                    {
                      className: "pause-button",
                      onClick: () => {
                        Ml().catch(
                          (E) => Rt(E.message)
                        );
                      },
                      children: [
                        /* @__PURE__ */ i.jsx(S1, { size: 13, fill: "currentColor" }),
                        " 暂停"
                      ]
                    }
                  )
                ] }) : /* @__PURE__ */ i.jsxs(
                  "button",
                  {
                    className: "record-button",
                    onClick: () => {
                      ra(nt);
                    },
                    children: [
                      /* @__PURE__ */ i.jsx(g1, { size: 15 }),
                      " 继续录音"
                    ]
                  }
                ),
                /* @__PURE__ */ i.jsx("button", { className: "icon-button", "aria-label": "更多", children: /* @__PURE__ */ i.jsx(m1, { size: 19 }) })
              ] })
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "mobile-tabs", children: [
              /* @__PURE__ */ i.jsx(
                "button",
                {
                  className: it === "discussion" ? "active" : "",
                  onClick: () => q("discussion"),
                  children: "讨论记录"
                }
              ),
              /* @__PURE__ */ i.jsxs(
                "button",
                {
                  className: it === "actions" ? "active" : "",
                  onClick: () => q("actions"),
                  children: [
                    "Action · ",
                    Yl.length
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "content-grid", children: [
              /* @__PURE__ */ i.jsx(
                V1,
                {
                  session: nt,
                  interimText: te === nt.id ? pt : "",
                  isRecording: Bl && te === nt.id,
                  search: $,
                  setSearch: el
                }
              ),
              /* @__PURE__ */ i.jsx(
                w1,
                {
                  session: nt,
                  actions: Yl,
                  runs: ft.runs,
                  settings: ft.settings,
                  onCreate: () => st("new"),
                  onExtract: () => wl(
                    () => Zt.extractActions(nt.id),
                    "已开始重新分析讨论"
                  ),
                  onEdit: (E) => st(E),
                  onDelete: (E) => {
                    window.confirm(`删除“${E.title}”？`) && wl(() => Zt.deleteAction(E.id));
                  },
                  onReady: (E) => wl(
                    () => Zt.updateAction(E.id, {
                      status: E.status === "ready" ? "draft" : "ready"
                    })
                  ),
                  onDispatch: (E) => wl(
                    () => Zt.dispatchAction(E.id),
                    "已把 Action 交给 Codex"
                  ),
                  onViewRun: (E) => M(E)
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs(
              "form",
              {
                className: "manual-discussion",
                onSubmit: async (E) => {
                  if (E.preventDefault(), !(!x.trim() || X)) {
                    W(!0);
                    try {
                      await Zt.addUtterance(nt.id, {
                        text: x.trim(),
                        atMs: Date.now() - Date.parse(nt.startedAt)
                      }), B(""), await il(!0);
                    } catch (Q) {
                      Rt(
                        Q instanceof Error ? Q.message : "记录失败"
                      );
                    } finally {
                      W(!1);
                    }
                  }
                },
                children: [
                  /* @__PURE__ */ i.jsx(
                    "textarea",
                    {
                      "aria-label": "补充讨论",
                      placeholder: "也可以打字补充讨论…",
                      maxLength: 4e3,
                      rows: 2,
                      value: x,
                      onChange: (E) => B(E.target.value)
                    }
                  ),
                  /* @__PURE__ */ i.jsx("button", { disabled: X || !x.trim(), children: "记录" }),
                  /* @__PURE__ */ i.jsx("small", { children: "录音保存在目标电脑。实时转写由浏览器提供，可能使用其在线识别服务。" })
                ]
              }
            )
          ] }) : /* @__PURE__ */ i.jsx(
            Z1,
            {
              project: zl,
              onStart: () => {
                Fu();
              },
              onProject: () => vl("new")
            }
          ) })
        ] }),
        Il && /* @__PURE__ */ i.jsx(
          k1,
          {
            settings: ft.settings,
            onClose: () => Bt(!1),
            onSave: async (E) => {
              try {
                await Zt.updateSettings(E), Bt(!1), await il(!0), Rt("系统设置已保存");
              } catch (Q) {
                Rt(
                  Q instanceof Error ? Q.message : "保存设置失败"
                );
              }
            }
          }
        ),
        Ot && /* @__PURE__ */ i.jsx(
          F1,
          {
            project: Ot === "new" ? void 0 : Ot,
            onClose: () => vl(null),
            onSave: async (E) => {
              try {
                const Q = Ot === "new" ? await Zt.createProject(E) : await Zt.updateProject(Ot.id, E);
                Z(Q.id), vl(null), await il(!0);
              } catch (Q) {
                Rt(
                  Q instanceof Error ? Q.message : "保存项目失败"
                );
              }
            }
          }
        ),
        p && /* @__PURE__ */ i.jsx(
          P1,
          {
            session: p,
            onClose: () => C(null),
            onSave: async (E) => {
              try {
                await Zt.updateSession(p.id, { title: E }), C(null), await il(!0);
              } catch (Q) {
                Rt(Q instanceof Error ? Q.message : "重命名失败");
              }
            }
          }
        ),
        V && nt && /* @__PURE__ */ i.jsx(
          I1,
          {
            action: V === "new" ? void 0 : V,
            onClose: () => st(null),
            onSave: async (E) => {
              await wl(
                () => V === "new" ? Zt.createAction(nt.id, E) : Zt.updateAction(V.id, E)
              ), st(null);
            }
          }
        ),
        d && /* @__PURE__ */ i.jsx(
          ty,
          {
            action: d,
            state: ft,
            onClose: () => M(null)
          }
        ),
        Y && /* @__PURE__ */ i.jsx(
          W1,
          {
            session: ft.sessions.find((E) => E.id === Y.id) || Y,
            onClose: () => H(null)
          }
        ),
        w && /* @__PURE__ */ i.jsxs("div", { className: "toast", children: [
          /* @__PURE__ */ i.jsx(Bd, { size: 16 }),
          " ",
          w
        ] })
      ]
    }
  );
}
function Z1({
  project: o,
  onStart: b,
  onProject: z
}) {
  return /* @__PURE__ */ i.jsxs("div", { className: "empty-workspace", children: [
    /* @__PURE__ */ i.jsxs("div", { className: "empty-orbit", children: [
      /* @__PURE__ */ i.jsx("span", { children: /* @__PURE__ */ i.jsx(Rf, { size: 28 }) }),
      /* @__PURE__ */ i.jsx("i", {}),
      /* @__PURE__ */ i.jsx("i", {}),
      /* @__PURE__ */ i.jsx("i", {})
    ] }),
    /* @__PURE__ */ i.jsx("p", { className: "overline", children: "CONVERSATION → ACTION" }),
    /* @__PURE__ */ i.jsxs("h1", { children: [
      "让讨论自然地",
      /* @__PURE__ */ i.jsx("br", {}),
      "变成下一步。"
    ] }),
    /* @__PURE__ */ i.jsxs("p", { className: "empty-lead", children: [
      "新建一个暂停中的会话，准备好后再继续录音。",
      /* @__PURE__ */ i.jsx("br", {}),
      "SaySo 会持续转写、分析，并把明确的事交给 Codex。"
    ] }),
    /* @__PURE__ */ i.jsxs("div", { className: "empty-actions", children: [
      /* @__PURE__ */ i.jsxs("button", { className: "primary-button", onClick: b, children: [
        /* @__PURE__ */ i.jsx(dc, { size: 17 }),
        " 新建会话"
      ] }),
      !o && /* @__PURE__ */ i.jsxs("button", { className: "ghost-button", onClick: z, children: [
        /* @__PURE__ */ i.jsx(Hf, { size: 16 }),
        " 添加项目"
      ] })
    ] })
  ] });
}
function L1({
  text: o,
  query: b
}) {
  return /* @__PURE__ */ i.jsx(i.Fragment, { children: U1(o, b).map(
    (z, r) => z.match ? /* @__PURE__ */ i.jsx(
      "mark",
      {
        className: "discussion-search-highlight",
        children: z.text
      },
      `${r}-${z.text}`
    ) : /* @__PURE__ */ i.jsx("span", { children: z.text }, `${r}-${z.text}`)
  ) });
}
function V1({
  session: o,
  interimText: b,
  isRecording: z,
  search: r,
  setSearch: U
}) {
  const D = R.useRef(null), _ = R.useRef(/* @__PURE__ */ new Map()), O = qf(r), N = R.useMemo(
    () => o.transcript.filter(
      (q) => _1(q.text, O)
    ),
    [o.transcript, O]
  ), x = R.useMemo(
    () => new Map(N.map((q, ot) => [q.id, ot])),
    [N]
  ), B = N.map((q) => q.id).join("|"), [X, W] = R.useState(-1);
  R.useEffect(() => {
    W(O && N.length ? 0 : -1);
  }, [O, o.id]), R.useEffect(() => {
    W(
      (q) => N.length === 0 ? -1 : Math.min(Math.max(q, 0), N.length - 1)
    );
  }, [N.length]), R.useEffect(() => {
    if (X < 0) return;
    _.current.get(N[X]?.id)?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [X, B, O]), R.useEffect(() => {
    O || D.current?.scrollIntoView({ behavior: "smooth" });
  }, [o.transcript.length, b, O]);
  const it = (q) => {
    N.length && W(
      (ot) => ot < 0 ? q === 1 ? 0 : N.length - 1 : (ot + q + N.length) % N.length
    );
  };
  return /* @__PURE__ */ i.jsxs("section", { className: "transcript-panel", children: [
    /* @__PURE__ */ i.jsxs("div", { className: "panel-header", children: [
      /* @__PURE__ */ i.jsxs("div", { children: [
        /* @__PURE__ */ i.jsx("h2", { children: "讨论记录" }),
        /* @__PURE__ */ i.jsxs("span", { children: [
          o.transcript.length,
          " 条"
        ] })
      ] }),
      /* @__PURE__ */ i.jsxs("div", { className: "transcript-tools", children: [
        /* @__PURE__ */ i.jsxs(
          "div",
          {
            className: `discussion-search ${O ? "has-query" : ""}`,
            children: [
              /* @__PURE__ */ i.jsxs("label", { className: "compact-search", children: [
                /* @__PURE__ */ i.jsx(p1, { size: 14 }),
                /* @__PURE__ */ i.jsx(
                  "input",
                  {
                    "aria-label": "搜索讨论",
                    value: r,
                    onChange: (q) => U(q.target.value),
                    onKeyDown: (q) => {
                      q.key === "Enter" && (q.preventDefault(), it(q.shiftKey ? -1 : 1)), q.key === "Escape" && U("");
                    },
                    placeholder: "搜索讨论"
                  }
                )
              ] }),
              O && /* @__PURE__ */ i.jsxs("div", { className: "search-navigation", children: [
                /* @__PURE__ */ i.jsx("span", { children: N.length ? `${X + 1}/${N.length}` : "0/0" }),
                /* @__PURE__ */ i.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !N.length,
                    onClick: () => it(-1),
                    "aria-label": "上一个搜索结果",
                    children: /* @__PURE__ */ i.jsx(d1, { size: 13 })
                  }
                ),
                /* @__PURE__ */ i.jsx(
                  "button",
                  {
                    type: "button",
                    disabled: !N.length,
                    onClick: () => it(1),
                    "aria-label": "下一个搜索结果",
                    children: /* @__PURE__ */ i.jsx(hc, { size: 13 })
                  }
                )
              ] })
            ]
          }
        ),
        /* @__PURE__ */ i.jsxs("span", { className: `live-chip ${z ? "on" : ""}`, children: [
          /* @__PURE__ */ i.jsx("i", {}),
          " ",
          z ? "录音中" : "已暂停"
        ] })
      ] })
    ] }),
    /* @__PURE__ */ i.jsx("div", { className: "transcript-scroll", children: o.transcript.length === 0 && !b ? /* @__PURE__ */ i.jsxs("div", { className: "transcript-empty", children: [
      /* @__PURE__ */ i.jsxs("div", { className: `sound-wave ${z ? "active" : ""}`, children: [
        /* @__PURE__ */ i.jsx("i", {}),
        /* @__PURE__ */ i.jsx("i", {}),
        /* @__PURE__ */ i.jsx("i", {}),
        /* @__PURE__ */ i.jsx("i", {}),
        /* @__PURE__ */ i.jsx("i", {})
      ] }),
      /* @__PURE__ */ i.jsx("h3", { children: z ? "正在听…" : "讨论已暂停" }),
      /* @__PURE__ */ i.jsx("p", { children: z ? "识别出的讨论会自动出现在这里。" : "点击右上角“继续录音”开始记录。" })
    ] }) : /* @__PURE__ */ i.jsxs("div", { className: "utterance-list", children: [
      o.transcript.map((q, ot) => {
        const Dt = Cd(q.createdAt), xt = ot === 0 || Dt !== Cd(o.transcript[ot - 1].createdAt), ft = G1(q.createdAt), Et = x.get(q.id), Lt = Et !== void 0 && Et === X;
        return /* @__PURE__ */ i.jsxs(
          "article",
          {
            className: `utterance ${Et !== void 0 ? "discussion-search-match" : ""} ${Lt ? "current-search-match" : ""}`,
            ref: (P) => {
              P ? _.current.set(q.id, P) : _.current.delete(q.id);
            },
            children: [
              /* @__PURE__ */ i.jsx(
                "time",
                {
                  className: xt ? "minute-mark" : void 0,
                  dateTime: q.createdAt,
                  "aria-label": xt ? `系统时间 ${ft}` : void 0,
                  "aria-hidden": !xt || void 0,
                  children: xt ? ft : null
                }
              ),
              /* @__PURE__ */ i.jsx("p", { children: /* @__PURE__ */ i.jsx(
                L1,
                {
                  text: q.text,
                  query: O
                }
              ) })
            ]
          },
          q.id
        );
      }),
      b && /* @__PURE__ */ i.jsxs("article", { className: "utterance interim", children: [
        /* @__PURE__ */ i.jsxs("time", { children: [
          /* @__PURE__ */ i.jsx(Yd, { size: 10 }),
          " LIVE"
        ] }),
        /* @__PURE__ */ i.jsxs("p", { children: [
          b,
          /* @__PURE__ */ i.jsx("span", { className: "typing-cursor" })
        ] })
      ] }),
      /* @__PURE__ */ i.jsx("div", { ref: D })
    ] }) })
  ] });
}
function w1({
  session: o,
  actions: b,
  runs: z,
  settings: r,
  onCreate: U,
  onExtract: D,
  onEdit: _,
  onDelete: O,
  onReady: N,
  onDispatch: x,
  onViewRun: B
}) {
  const X = ["queued", "running"].includes(
    o.analysisStatus || ""
  ), [W, it] = R.useState(!0), [q, ot] = R.useState(!1), xt = o.analysisTraces?.find(
    (Z) => Z.id === o.analysisActiveTraceId
  ) || o.analysisTraces?.[0], ft = R.useRef(void 0);
  R.useEffect(() => {
    X && xt?.id && ft.current !== xt.id && (ft.current = xt.id, it(!0));
  }, [X, xt?.id]);
  const Et = [
    "读取最近 5 分钟的原始讨论",
    `对照 ${b.filter((Z) => !["running", "done"].includes(Z.status)).length} 条已有需求`,
    `${r.analysisModel === "default" ? "Codex" : r.analysisModel} 正在判断补充与修正`
  ], Lt = xt && xt.events.length >= 3 ? xt.events.slice(-3) : Et, P = Qd(
    Lt,
    X && W
  ), Vt = [...P].reverse().find(Boolean);
  return /* @__PURE__ */ i.jsxs("aside", { className: "action-panel", children: [
    /* @__PURE__ */ i.jsxs("div", { className: "action-heading", children: [
      /* @__PURE__ */ i.jsxs("div", { children: [
        /* @__PURE__ */ i.jsxs("span", { className: "action-kicker", children: [
          /* @__PURE__ */ i.jsx(z1, { size: 13, fill: "currentColor" }),
          " ACTIONS"
        ] }),
        /* @__PURE__ */ i.jsx("h2", { children: "接下来要做的事" })
      ] }),
      /* @__PURE__ */ i.jsx(
        "button",
        {
          className: "icon-button bordered",
          onClick: U,
          "aria-label": "添加 Action",
          children: /* @__PURE__ */ i.jsx(dc, { size: 18 })
        }
      )
    ] }),
    /* @__PURE__ */ i.jsxs(
      "div",
      {
        className: `analysis-strip ${o.analysisStatus === "failed" ? "failed" : ""} ${W ? "expanded" : ""}`,
        children: [
          X ? /* @__PURE__ */ i.jsx(me, { size: 13, className: "spin" }) : /* @__PURE__ */ i.jsx(_d, { size: 13 }),
          /* @__PURE__ */ i.jsxs(
            "button",
            {
              className: "analysis-status-copy",
              onClick: () => X && it((Z) => !Z),
              disabled: !X,
              children: [
                /* @__PURE__ */ i.jsx("span", { children: X ? W && Vt || "分析中" : o.analysisStatus === "failed" ? o.analysisError || "分析失败，点击重试" : o.status === "recording" ? "每 2 分钟分析最近 5 分钟讨论" : `${b.length} 条 Action` }),
                X && (W ? /* @__PURE__ */ i.jsx(hc, { size: 12 }) : /* @__PURE__ */ i.jsx(Uf, { size: 12 }))
              ]
            }
          ),
          xt && /* @__PURE__ */ i.jsx(
            "button",
            {
              className: "analysis-details-button",
              onClick: () => ot(!0),
              children: "详情"
            }
          ),
          /* @__PURE__ */ i.jsx(
            "button",
            {
              className: "analysis-refresh",
              onClick: D,
              title: "重新分析讨论",
              children: /* @__PURE__ */ i.jsx(x1, { size: 13 })
            }
          ),
          X && W && /* @__PURE__ */ i.jsx("div", { className: "analysis-detail-lines", children: Lt.map((Z, zt) => /* @__PURE__ */ i.jsxs(
            "div",
            {
              className: P[zt] === Z ? "complete" : P[zt] ? "typing" : "",
              children: [
                /* @__PURE__ */ i.jsx("span", { children: zt + 1 }),
                /* @__PURE__ */ i.jsx("p", { children: P[zt] || " " })
              ]
            },
            `${zt}-${Z}`
          )) })
        ]
      }
    ),
    /* @__PURE__ */ i.jsx("div", { className: "action-list", children: b.length === 0 ? /* @__PURE__ */ i.jsxs("div", { className: "action-empty-compact", children: [
      /* @__PURE__ */ i.jsx(_d, { size: 17 }),
      /* @__PURE__ */ i.jsx("span", { children: X ? "正在整理讨论…" : "还没有明确的 Action" })
    ] }) : b.map((Z) => {
      const zt = z.find(($t) => $t.id === Z.runId);
      return /* @__PURE__ */ i.jsxs(
        "article",
        {
          className: `action-card card-${Z.status}`,
          children: [
            /* @__PURE__ */ i.jsxs("div", { className: "action-card-top", children: [
              /* @__PURE__ */ i.jsx(Xd, { status: Z.status }),
              /* @__PURE__ */ i.jsx("span", { className: `priority priority-${Z.priority}`, children: Z.priority === "high" ? "高优先级" : Z.priority === "low" ? "低优先级" : "普通" }),
              /* @__PURE__ */ i.jsxs("div", { className: "card-tools", children: [
                !["running", "done"].includes(Z.status) && /* @__PURE__ */ i.jsx("button", { onClick: () => _(Z), "aria-label": "编辑", children: /* @__PURE__ */ i.jsx(oc, { size: 14 }) }),
                !["running", "done"].includes(Z.status) && /* @__PURE__ */ i.jsx(
                  "button",
                  {
                    onClick: () => O(Z),
                    "aria-label": "删除",
                    children: /* @__PURE__ */ i.jsx(A1, { size: 14 })
                  }
                )
              ] })
            ] }),
            /* @__PURE__ */ i.jsx(
              K1,
              {
                action: Z,
                trace: xt,
                active: X && !["running", "done"].includes(Z.status)
              }
            ),
            /* @__PURE__ */ i.jsx("h3", { children: Z.title }),
            Z.detail && /* @__PURE__ */ i.jsx("p", { className: "action-detail", children: Z.detail }),
            /* @__PURE__ */ i.jsxs("blockquote", { children: [
              /* @__PURE__ */ i.jsx("span", { children: "“" }),
              Z.source
            ] }),
            Z.status === "running" && /* @__PURE__ */ i.jsxs(
              "button",
              {
                className: "run-progress",
                onClick: () => B(Z),
                children: [
                  /* @__PURE__ */ i.jsxs("span", { children: [
                    /* @__PURE__ */ i.jsx(me, { size: 15 }),
                    " Codex 正在项目中工作"
                  ] }),
                  /* @__PURE__ */ i.jsx("small", { children: "查看输出" })
                ]
              }
            ),
            Z.status === "done" && /* @__PURE__ */ i.jsxs(
              "button",
              {
                className: "run-complete",
                onClick: () => B(Z),
                children: [
                  /* @__PURE__ */ i.jsxs("span", { children: [
                    /* @__PURE__ */ i.jsx(Bd, { size: 15 }),
                    " 已由 Codex 完成"
                  ] }),
                  /* @__PURE__ */ i.jsx("small", { children: "查看结果" })
                ]
              }
            ),
            Z.status === "failed" && zt?.error && /* @__PURE__ */ i.jsx("p", { className: "run-error", children: zt.error }),
            !["running", "done"].includes(Z.status) && /* @__PURE__ */ i.jsxs("div", { className: "action-card-footer", children: [
              /* @__PURE__ */ i.jsxs(
                "button",
                {
                  className: `ready-toggle ${Z.status === "ready" ? "checked" : ""}`,
                  onClick: () => N(Z),
                  children: [
                    /* @__PURE__ */ i.jsx("span", { children: Z.status === "ready" && /* @__PURE__ */ i.jsx($u, { size: 11 }) }),
                    " ",
                    "讨论完善"
                  ]
                }
              ),
              /* @__PURE__ */ i.jsxs(
                "button",
                {
                  className: "dispatch-button",
                  disabled: Z.status === "draft",
                  onClick: () => x(Z),
                  children: [
                    /* @__PURE__ */ i.jsx(j1, { size: 14 }),
                    " ",
                    Z.status === "failed" ? "重新交给 Codex" : "交给 Codex"
                  ]
                }
              )
            ] })
          ]
        },
        Z.id
      );
    }) }),
    /* @__PURE__ */ i.jsxs("div", { className: "action-footnote", children: [
      /* @__PURE__ */ i.jsx(v1, { size: 13 }),
      /* @__PURE__ */ i.jsx("span", { children: "Codex 将在当前项目目录内执行" })
    ] }),
    q && xt && /* @__PURE__ */ i.jsx(
      J1,
      {
        trace: xt,
        onClose: () => ot(!1)
      }
    )
  ] });
}
function K1({
  action: o,
  trace: b,
  active: z
}) {
  const [r, U] = R.useState(!0), D = R.useRef(void 0);
  R.useEffect(() => {
    z && b?.id && D.current !== b.id && (D.current = b.id, U(!0));
  }, [z, b?.id]);
  const _ = [
    `输入：核对${b?.contextLabel || "最近 5 分钟"}的 ${b?.utteranceCount || 0} 条讨论`,
    `判断“${o.title}”是否需要补充或修正`
  ], O = Qd(_, z && r);
  return z ? /* @__PURE__ */ i.jsxs("div", { className: `requirement-analysis ${r ? "expanded" : ""}`, children: [
    /* @__PURE__ */ i.jsxs(
      "button",
      {
        className: "requirement-analysis-toggle",
        onClick: () => U((N) => !N),
        children: [
          /* @__PURE__ */ i.jsx(me, { size: 11, className: "spin" }),
          /* @__PURE__ */ i.jsx("span", { children: r ? "需求分析中" : "分析中" }),
          r ? /* @__PURE__ */ i.jsx(hc, { size: 11 }) : /* @__PURE__ */ i.jsx(Uf, { size: 11 })
        ]
      }
    ),
    r && /* @__PURE__ */ i.jsx("div", { className: "requirement-analysis-lines", children: _.map((N, x) => /* @__PURE__ */ i.jsx(
      "p",
      {
        className: O[x] === N ? "complete" : O[x] ? "typing" : "",
        children: O[x] || " "
      },
      N
    )) })
  ] }) : null;
}
function J1({
  trace: o,
  onClose: b
}) {
  const z = R.useRef(null);
  R.useEffect(() => {
    z.current?.scrollIntoView({
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
  return /* @__PURE__ */ i.jsxs(
    ia,
    {
      wide: !0,
      title: "分析详情",
      description: `${o.model} · ${r} · ${o.utteranceCount} 条原始讨论`,
      onClose: b,
      children: [
        /* @__PURE__ */ i.jsxs("div", { className: "analysis-trace-scroll", children: [
          /* @__PURE__ */ i.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ i.jsxs("div", { className: "analysis-trace-heading", children: [
              /* @__PURE__ */ i.jsx("h3", { children: "实时过程" }),
              /* @__PURE__ */ i.jsxs("span", { className: `trace-status trace-${o.status}`, children: [
                o.status === "running" && /* @__PURE__ */ i.jsx(me, { size: 11, className: "spin" }),
                r
              ] })
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "analysis-live-log", children: [
              o.events.map((U, D) => /* @__PURE__ */ i.jsxs("div", { children: [
                /* @__PURE__ */ i.jsx("span", { children: String(D + 1).padStart(2, "0") }),
                /* @__PURE__ */ i.jsx("p", { children: U })
              ] }, `${D}-${U}`)),
              /* @__PURE__ */ i.jsx("div", { ref: z })
            ] })
          ] }),
          /* @__PURE__ */ i.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ i.jsx("h3", { children: "原始输入" }),
            /* @__PURE__ */ i.jsx("pre", { children: o.input || "暂无输入" })
          ] }),
          /* @__PURE__ */ i.jsxs("section", { className: "analysis-trace-section", children: [
            /* @__PURE__ */ i.jsx("h3", { children: "最终输出" }),
            /* @__PURE__ */ i.jsx("pre", { children: o.output || (o.status === "running" ? "等待模型返回最终结构化输出…" : o.error || "本轮没有输出") })
          ] })
        ] }),
        /* @__PURE__ */ i.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ i.jsx("button", { className: "primary-button", onClick: b, children: "关闭" }) })
      ]
    }
  );
}
function k1({
  settings: o,
  onClose: b,
  onSave: z
}) {
  const [r, U] = R.useState(o), [D, _] = R.useState(!1);
  return /* @__PURE__ */ i.jsx(
    ia,
    {
      title: "系统设置",
      description: "配置讨论如何被分析，以及 Action 由谁执行。",
      onClose: b,
      children: /* @__PURE__ */ i.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), _(!0), z(r).finally(() => _(!1));
          },
          children: [
            /* @__PURE__ */ i.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ i.jsx("strong", { children: "讨论分析" }),
                /* @__PURE__ */ i.jsx("small", { children: "从转写中持续提炼决定和待办" })
              ] }),
              /* @__PURE__ */ i.jsxs(
                "select",
                {
                  value: r.analysisProvider,
                  onChange: (O) => U({
                    ...r,
                    analysisProvider: O.target.value
                  }),
                  children: [
                    /* @__PURE__ */ i.jsx("option", { value: "codex", children: "Codex" }),
                    /* @__PURE__ */ i.jsx("option", { value: "local", children: "本地规则" })
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ i.jsx("strong", { children: "分析模型" }),
                /* @__PURE__ */ i.jsx("small", { children: "小模型更快，也更节省额度" })
              ] }),
              /* @__PURE__ */ i.jsx(
                "select",
                {
                  disabled: r.analysisProvider === "local",
                  value: r.analysisModel,
                  onChange: (O) => U({ ...r, analysisModel: O.target.value }),
                  children: B1.map((O) => /* @__PURE__ */ i.jsx("option", { value: O.value, children: O.label }, O.value))
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ i.jsx("strong", { children: "持续分析" }),
                /* @__PURE__ */ i.jsx("small", { children: "录音中每 2 分钟运行一次，每次读取最近 5 分钟讨论" })
              ] }),
              /* @__PURE__ */ i.jsx("span", { className: "setting-fixed-value", children: "2 分钟 / 5 分钟窗口" })
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ i.jsx("strong", { children: "执行模型" }),
                /* @__PURE__ */ i.jsx("small", { children: "最终进入项目实现 Action 时使用" })
              ] }),
              /* @__PURE__ */ i.jsx(
                "select",
                {
                  value: r.executionModel,
                  onChange: (O) => U({ ...r, executionModel: O.target.value }),
                  children: Y1.map((O) => /* @__PURE__ */ i.jsx("option", { value: O.value, children: O.label }, O.value))
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "setting-group", children: [
              /* @__PURE__ */ i.jsxs("div", { className: "setting-copy", children: [
                /* @__PURE__ */ i.jsx("strong", { children: "Action 执行器" }),
                /* @__PURE__ */ i.jsx("small", { children: "在项目目录内完成已确认的工作" })
              ] }),
              /* @__PURE__ */ i.jsx("select", { value: "codex", disabled: !0, children: /* @__PURE__ */ i.jsx("option", { value: "codex", children: "Codex" }) })
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ i.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ i.jsxs("button", { className: "primary-button", disabled: D, children: [
                D ? /* @__PURE__ */ i.jsx(me, { size: 15 }) : /* @__PURE__ */ i.jsx($u, { size: 15 }),
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
  recording: o,
  index: b
}) {
  const [z, r] = R.useState(""), [U, D] = R.useState(""), [_, O] = R.useState(!1), N = R.useRef(null);
  R.useEffect(
    () => () => {
      N.current?.abort();
    },
    []
  ), R.useEffect(
    () => () => {
      z && URL.revokeObjectURL(z);
    },
    [z]
  );
  const x = async () => {
    O(!0), D("");
    const q = new AbortController();
    N.current = q;
    try {
      r(await R1(o.id, q.signal));
    } catch (ot) {
      q.signal.aborted || D(ot instanceof Error ? ot.message : "无法读取录音");
    } finally {
      q.signal.aborted || O(!1);
    }
  }, [B, X] = R.useState(), W = o.durationMs ?? B, it = (q) => {
    if (Number.isFinite(q.duration) && q.duration > 0) {
      X(q.duration * 1e3), q.dataset.durationProbe === "active" && (q.currentTime = 0, delete q.dataset.durationProbe);
      return;
    }
    o.durationMs === void 0 && q.dataset.durationProbe !== "active" && (q.dataset.durationProbe = "active", q.currentTime = Number.MAX_SAFE_INTEGER);
  };
  return /* @__PURE__ */ i.jsxs("article", { className: "recording-row", children: [
    /* @__PURE__ */ i.jsxs("div", { className: "recording-index", children: [
      /* @__PURE__ */ i.jsx(Cf, { size: 15 }),
      /* @__PURE__ */ i.jsx("span", { children: String(b + 1).padStart(2, "0") })
    ] }),
    /* @__PURE__ */ i.jsxs("div", { className: "recording-info", children: [
      /* @__PURE__ */ i.jsxs("div", { children: [
        /* @__PURE__ */ i.jsxs("strong", { children: [
          "录音片段 ",
          b + 1
        ] }),
        /* @__PURE__ */ i.jsx("span", { children: new Date(o.startedAt).toLocaleString("zh-CN", {
          month: "numeric",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit"
        }) })
      ] }),
      /* @__PURE__ */ i.jsxs("small", { children: [
        W !== void 0 ? Gd(W) : "读取时长中",
        " ·",
        " ",
        X1(o.sizeBytes)
      ] }),
      !z && /* @__PURE__ */ i.jsx(
        "button",
        {
          className: "ghost-button",
          disabled: _,
          onClick: () => {
            x();
          },
          children: _ ? "正在从设备读取…" : "加载录音"
        }
      ),
      U && /* @__PURE__ */ i.jsx("p", { role: "alert", children: U }),
      z && /* @__PURE__ */ i.jsx(
        "audio",
        {
          controls: !0,
          preload: "metadata",
          src: z,
          onLoadedMetadata: (q) => it(q.currentTarget),
          onDurationChange: (q) => it(q.currentTarget),
          onSeeked: (q) => it(q.currentTarget),
          onPlay: (q) => {
            q.currentTarget.getRootNode().querySelectorAll("audio").forEach((ot) => {
              ot !== q.currentTarget && ot.pause();
            });
          }
        }
      )
    ] })
  ] });
}
function W1({
  session: o,
  onClose: b
}) {
  const z = [...o.recordings || []].sort(
    (r, U) => r.startedAt.localeCompare(U.startedAt)
  );
  return /* @__PURE__ */ i.jsxs(
    ia,
    {
      wide: !0,
      title: "录音片段",
      description: `${o.title} · 共 ${z.length} 段，每段最长 10 分钟`,
      onClose: b,
      children: [
        /* @__PURE__ */ i.jsx("div", { className: "recordings-list", children: z.length ? z.map((r, U) => /* @__PURE__ */ i.jsx(
          $1,
          {
            recording: r,
            index: U
          },
          r.id
        )) : /* @__PURE__ */ i.jsxs("div", { className: "recordings-empty", children: [
          /* @__PURE__ */ i.jsx(Cf, { size: 20 }),
          "暂时没有录音"
        ] }) }),
        /* @__PURE__ */ i.jsx("div", { className: "modal-actions", children: /* @__PURE__ */ i.jsx("button", { className: "primary-button", onClick: b, children: "关闭" }) })
      ]
    }
  );
}
function F1({
  project: o,
  onClose: b,
  onSave: z
}) {
  const [r, U] = R.useState(o?.name || ""), [D, _] = R.useState(o?.rootPath || ""), [O, N] = R.useState(!1);
  return /* @__PURE__ */ i.jsx(
    ia,
    {
      title: o ? "编辑项目" : "添加项目",
      description: "Action 会在这个本地目录中交给 Codex。",
      onClose: b,
      children: /* @__PURE__ */ i.jsxs(
        "form",
        {
          onSubmit: (x) => {
            x.preventDefault(), N(!0), z({ name: r, rootPath: D }).finally(() => N(!1));
          },
          children: [
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "项目名称" }),
            /* @__PURE__ */ i.jsx(
              "input",
              {
                className: "text-field",
                value: r,
                onChange: (x) => U(x.target.value),
                placeholder: "我的产品"
              }
            ),
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "本地项目目录" }),
            /* @__PURE__ */ i.jsxs("div", { className: "path-field", children: [
              /* @__PURE__ */ i.jsx(Hf, { size: 16 }),
              /* @__PURE__ */ i.jsx(
                "input",
                {
                  value: D,
                  onChange: (x) => _(x.target.value),
                  placeholder: "C:\\\\Code\\\\my-project"
                }
              )
            ] }),
            /* @__PURE__ */ i.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ i.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ i.jsxs(
                "button",
                {
                  className: "primary-button",
                  disabled: O || !r.trim() || !D.trim(),
                  children: [
                    O ? /* @__PURE__ */ i.jsx(me, { size: 15 }) : /* @__PURE__ */ i.jsx($u, { size: 15 }),
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
  session: o,
  onClose: b,
  onSave: z
}) {
  const [r, U] = R.useState(o.title), [D, _] = R.useState(!1);
  return /* @__PURE__ */ i.jsx(
    ia,
    {
      title: "重命名会话",
      description: "日期时间名称也可以随时改掉。",
      onClose: b,
      children: /* @__PURE__ */ i.jsxs(
        "form",
        {
          onSubmit: (O) => {
            O.preventDefault(), _(!0), z(r).finally(() => _(!1));
          },
          children: [
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "会话名称" }),
            /* @__PURE__ */ i.jsx(
              "input",
              {
                className: "text-field",
                autoFocus: !0,
                value: r,
                onChange: (O) => U(O.target.value)
              }
            ),
            /* @__PURE__ */ i.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ i.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ i.jsxs("button", { className: "primary-button", disabled: D || !r.trim(), children: [
                D ? /* @__PURE__ */ i.jsx(me, { size: 15 }) : /* @__PURE__ */ i.jsx($u, { size: 15 }),
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
  action: o,
  onClose: b,
  onSave: z
}) {
  const [r, U] = R.useState(o?.title || ""), [D, _] = R.useState(o?.detail || ""), [O, N] = R.useState(
    o?.priority || "medium"
  ), [x, B] = R.useState(!1);
  return /* @__PURE__ */ i.jsx(
    ia,
    {
      title: o ? "完善 Action" : "添加 Action",
      description: "把目标和验收结果说清楚，Codex 会做得更稳。",
      onClose: b,
      children: /* @__PURE__ */ i.jsxs(
        "form",
        {
          onSubmit: (X) => {
            X.preventDefault(), B(!0), z({ title: r, detail: D, priority: O }).finally(
              () => B(!1)
            );
          },
          children: [
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "要做什么" }),
            /* @__PURE__ */ i.jsx(
              "input",
              {
                className: "text-field",
                value: r,
                autoFocus: !0,
                onChange: (X) => U(X.target.value),
                placeholder: "一句话描述结果"
              }
            ),
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "详细说明" }),
            /* @__PURE__ */ i.jsx(
              "textarea",
              {
                className: "text-field textarea",
                value: D,
                onChange: (X) => _(X.target.value),
                placeholder: "补充范围、约束和验收标准…"
              }
            ),
            /* @__PURE__ */ i.jsx("label", { className: "field-label", children: "优先级" }),
            /* @__PURE__ */ i.jsx("div", { className: "priority-picker", children: ["high", "medium", "low"].map((X) => /* @__PURE__ */ i.jsxs(
              "button",
              {
                type: "button",
                className: O === X ? "active" : "",
                onClick: () => N(X),
                children: [
                  /* @__PURE__ */ i.jsx(h1, { size: 9, fill: "currentColor" }),
                  " ",
                  X === "high" ? "高" : X === "medium" ? "普通" : "低"
                ]
              },
              X
            )) }),
            /* @__PURE__ */ i.jsxs("div", { className: "modal-actions", children: [
              /* @__PURE__ */ i.jsx("button", { type: "button", className: "ghost-button", onClick: b, children: "取消" }),
              /* @__PURE__ */ i.jsxs("button", { className: "primary-button", disabled: x || !r.trim(), children: [
                x ? /* @__PURE__ */ i.jsx(me, { size: 15 }) : /* @__PURE__ */ i.jsx($u, { size: 15 }),
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
  action: o,
  state: b,
  onClose: z
}) {
  const [r, U] = R.useState("");
  o = b.actions.find((_) => _.id === o.id) || o;
  const D = b.runs.find((_) => _.id === o.runId);
  return /* @__PURE__ */ i.jsxs(
    ia,
    {
      wide: !0,
      title: o.status === "running" ? "Codex 正在工作" : o.status === "done" ? "Codex 已完成" : "Codex 执行结果",
      description: o.title,
      onClose: z,
      children: [
        /* @__PURE__ */ i.jsxs("div", { className: "run-meta", children: [
          /* @__PURE__ */ i.jsx(Xd, { status: o.status }),
          D && /* @__PURE__ */ i.jsxs("span", { children: [
            /* @__PURE__ */ i.jsx(y1, { size: 13 }),
            " ",
            new Date(D.startedAt).toLocaleString("zh-CN")
          ] })
        ] }),
        /* @__PURE__ */ i.jsx("pre", { className: "run-output", children: D?.output || (D?.status === "running" ? "正在启动 Codex…" : D?.error || "暂无输出") }),
        D?.error && /* @__PURE__ */ i.jsx("p", { role: "alert", children: D.error }),
        r && /* @__PURE__ */ i.jsx("p", { role: "alert", children: r }),
        /* @__PURE__ */ i.jsxs("div", { className: "modal-actions", children: [
          D?.status === "running" && /* @__PURE__ */ i.jsx(
            "button",
            {
              className: "ghost-button",
              onClick: () => {
                Zt.cancelRun(D.id).catch((_) => U(_.message));
              },
              children: "停止执行"
            }
          ),
          /* @__PURE__ */ i.jsx("button", { className: "primary-button", onClick: z, children: "关闭" })
        ] })
      ]
    }
  );
}
function ly(o, b, z, r) {
  M1(b);
  const U = o.shadowRoot || o.attachShadow({ mode: "open" });
  U.replaceChildren();
  const D = document.createElement("link");
  D.rel = "stylesheet", D.href = new URL(
    /* @vite-ignore */
    "./sayso.css",
    import.meta.url
  ).href;
  const _ = document.createElement("div");
  _.id = "root", U.append(D, _);
  const O = f1.createRoot(_);
  return O.render(/* @__PURE__ */ i.jsx(Q1, { onClose: z, deviceLabel: r })), () => O.unmount();
}
export {
  ly as mount
};
