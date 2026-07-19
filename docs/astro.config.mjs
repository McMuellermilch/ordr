import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://McMuellermilch.github.io",
  base: "/ordr",
  integrations: [
    starlight({
      title: "ordr",
      tagline: "File chaos, organized.",
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
          link: "getting-started",
        },
        {
          label: "Commands",
          collapsed: false,
          autogenerate: { directory: "commands" },
        },
        {
          label: "Configuration",
          collapsed: false,
          autogenerate: { directory: "config" },
        },
        {
          label: "Guides",
          collapsed: false,
          autogenerate: { directory: "guides" },
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
