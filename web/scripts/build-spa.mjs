import { spawn } from "node:child_process"

process.env.BBSGO_WEB_SPA = "true"

const build = spawn("pnpm exec react-router build", {
  stdio: "inherit",
  env: process.env,
  shell: true,
})

build.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal)
    return
  }
  process.exit(code ?? 1)
})
