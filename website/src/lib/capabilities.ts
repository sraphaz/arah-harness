import { readFileSync } from "fs";
import path from "path";
import { load as loadYaml } from "js-yaml";

export type Capabilities = {
  version: number;
  harness_version: string;
  available: Array<{ id: string; name: string; notes?: string }>;
  experimental: Array<{ id: string; name: string; notes?: string }>;
  planned: Array<{ id: string; name: string; notes?: string }>;
};

export function loadCapabilities(): Capabilities {
  const candidates = [
    path.join(process.cwd(), "..", "capabilities.yaml"),
    path.join(process.cwd(), "capabilities.yaml"),
  ];
  for (const p of candidates) {
    try {
      const raw = readFileSync(p, "utf8");
      return loadYaml(raw) as Capabilities;
    } catch {
      /* try next */
    }
  }
  return {
    version: 1,
    harness_version: "0.3.1",
    available: [],
    experimental: [],
    planned: [],
  };
}
