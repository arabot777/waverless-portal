import { createContext, useContext, useEffect, useState, useCallback, useRef } from "react"

export type Theme = "dark" | "light" | "system"

type ThemeContextType = {
  theme: Theme
  setTheme: (theme: Theme) => void
}

const ThemeContext = createContext<ThemeContextType>({ theme: "system", setTheme: () => {} })

function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) return parts.pop()?.split(';').shift() || null
  return null
}

function getInitialTheme(): Theme {
  const cookieTheme = getCookie('theme')
  if (cookieTheme === 'dark' || cookieTheme === 'light') return cookieTheme
  return 'system'
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>(getInitialTheme)
  const previousCookieTheme = useRef<string | null>(null)

  const applyTheme = useCallback((t: Theme) => {
    const root = document.documentElement
    root.classList.remove("light", "dark")
    const resolved = t === "system" 
      ? (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light")
      : t
    root.classList.add(resolved)
  }, [])

  useEffect(() => { applyTheme(theme) }, [theme, applyTheme])

  useEffect(() => {
    const checkCookie = () => {
      const cookieTheme = getCookie('theme')
      if (cookieTheme !== previousCookieTheme.current) {
        previousCookieTheme.current = cookieTheme
        if (cookieTheme === 'dark' || cookieTheme === 'light') setTheme(cookieTheme)
        else setTheme('system')
      }
    }
    previousCookieTheme.current = getCookie('theme')
    const id = setInterval(checkCookie, 500)
    return () => clearInterval(id)
  }, [])

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export const useTheme = () => useContext(ThemeContext)
