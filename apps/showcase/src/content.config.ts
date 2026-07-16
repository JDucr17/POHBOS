import { defineCollection } from "astro:content";
import { glob } from "astro/loaders";
import { z } from "astro/zod";

const docs = defineCollection({
  loader: glob({ pattern: "**/*.mdx", base: "./src/content/docs" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    accent: z.enum(["primary", "flow", "cold"]).default("primary"),
    section: z.enum(["Concepts", "Services"]).default("Services"),
  }),
});

export const collections = { docs };
