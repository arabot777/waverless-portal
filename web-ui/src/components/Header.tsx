interface HeaderProps {
  title?: string
}

export default function Header({ title = 'Dashboard' }: HeaderProps) {
  return (
    <header className="header">
      <h1 className="page-title">{title}</h1>
    </header>
  )
}
