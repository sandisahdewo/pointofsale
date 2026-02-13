import './globals.css';

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <title>Point of Sale - Admin Panel</title>
        <meta name="description" content="Point of Sale Admin Panel" />
      </head>
      <body>{children}</body>
    </html>
  );
}
