import js from "@eslint/js";
import { defineConfig } from "eslint/config";
import ts from "typescript-eslint";
import astro from "eslint-plugin-astro";
import svelte from "eslint-plugin-svelte";
import prettier from "eslint-config-prettier/flat";
import globals from "globals";

export default defineConfig(
  {
    ignores: ["dist/", ".astro/", "node_modules/"],
  },

  js.configs.recommended,
  ts.configs.recommended,
  ...astro.configs.recommended,
  svelte.configs.recommended,

  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },

  {
    files: ["**/*.svelte", "**/*.svelte.ts", "**/*.svelte.js"],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: [".svelte"],
        parser: ts.parser,
      },
    },
  },

  {
    rules: {
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },

  prettier,
  svelte.configs.prettier,
);