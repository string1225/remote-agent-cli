import { createRoot } from "react-dom/client";
import App from "./App";
import { configure, type Transport } from "./transport";
import "./styles.css";

export function mount(
  host: HTMLElement,
  transport: Transport,
  onClose: () => void,
  label: string,
) {
  configure(transport);
  const shadow = host.shadowRoot || host.attachShadow({ mode: "open" });
  shadow.replaceChildren();
  const link = document.createElement("link");
  link.rel = "stylesheet";
  link.href = new URL(/* @vite-ignore */ "./sayso.css", import.meta.url).href;
  const container = document.createElement("div");
  container.id = "root";
  shadow.append(link, container);
  const root = createRoot(container);
  root.render(<App onClose={onClose} deviceLabel={label} />);
  return () => root.unmount();
}
