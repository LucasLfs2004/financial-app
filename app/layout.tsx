import type { Metadata, Viewport } from "next";
import "./globals.css";
import { PwaRegister } from "./pwa-register";

export const metadata: Metadata = {
  title: {
    default: "Projeção",
    template: "%s · Projeção",
  },
  description: "Planejamento financeiro simples, visual e sob seu controle.",
  applicationName: "Projeção",
};

export const viewport: Viewport = {
  themeColor: "#07111f",
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="pt-BR" className="dark" data-scroll-behavior="smooth">
      <body>
        <PwaRegister />
        {children}
      </body>
    </html>
  );
}
