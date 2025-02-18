import { extendTheme } from '@chakra-ui/react'

type ThemeConfigType = {
  initialColorMode: 'light' | 'dark' | 'system'
  useSystemColorMode: boolean
}

const config: ThemeConfigType = {
  initialColorMode: 'system',
  useSystemColorMode: true,
}

const theme = extendTheme({ config })

export default theme 