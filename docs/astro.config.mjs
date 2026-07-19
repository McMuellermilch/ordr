import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://McMuellermilch.github.io",
  base: "/ordr",
  integrations: [
    starlight({
      title: "ordr",
      tagline: "File chaos, organized.",
      logo: {
        light: "./src/assets/logo-light.svg",
        dark: "./src/assets/logo-dark.svg",
        replacesTitle: false,
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/McMuellermilch/ordr",
        },
      ],
      customCss: ["./src/styles/custom.css"],
      defaultLocale: "en",
      sidebar: [
        {
          label: "Getting Started",
          link: "/getting-started",
        },
        {
          label: "Commands",
          items: [
            { label: "clean", link: "/commands/clean" },
            { label: "preview", link: "/commands/preview" },
            { label: "undo", link: "/commands/undo" },
            { label: "rules", link: "/commands/rules" },
            { label: "init", link: "/commands/init" },
            { label: "status", link: "/commands/status" },
          ],
        },
        {
          label: "Configuration",
          items: [
            { label: "Overview", link: "/config/overview" },
            { label: "Rule Types & Matchers", link: "/config/rules" },
            { label: "Examples", link: "/config/examples" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Install via Homebrew", link: "/guides/homebrew" },
            {
              label: "Migrating from maid",
              link: "/guides/migrating-from-maid",
            },
          ],
        },
      ],
      head: [
        {
          tag: "meta",
          attrs: {
            name: "theme-color",
            content: "#7c3aed",
          },
        },
      ],
    }),
  ],
});
