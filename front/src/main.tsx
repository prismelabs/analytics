import { render } from "preact";
import { App } from "@/pages/app.tsx";
import "@/styles/style.css";

render(<App />, document.getElementById("app") as HTMLElement);
