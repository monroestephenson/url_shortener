import {
  Box,
  Flex,
  Button,
  Heading,
} from '@chakra-ui/react'
import { useColorMode, useColorModeValue } from '@chakra-ui/system'
import { Link } from 'react-router-dom'
import { MoonIcon, SunIcon } from '@chakra-ui/icons'
import { useAuth } from '../contexts/AuthContext'

const Navbar = () => {
  const { colorMode, toggleColorMode } = useColorMode()
  const { isAuthenticated, logout } = useAuth()
  const bg = useColorModeValue('white', 'gray.800')

  return (
    <Box bg={bg} px={4} shadow="sm">
      <Flex h={16} alignItems="center" justifyContent="space-between">
        <Link to="/">
          <Heading size="md">URL Shortener</Heading>
        </Link>

        <Flex alignItems="center" gap={4}>
          <Button onClick={toggleColorMode}>
            {colorMode === 'light' ? <MoonIcon /> : <SunIcon />}
          </Button>

          {isAuthenticated ? (
            <Button onClick={logout} variant="ghost">
              Logout
            </Button>
          ) : (
            <>
              <Button as={Link} to="/login" variant="ghost">
                Login
              </Button>
              <Button as={Link} to="/register" colorScheme="blue">
                Register
              </Button>
            </>
          )}
        </Flex>
      </Flex>
    </Box>
  )
}

export default Navbar 