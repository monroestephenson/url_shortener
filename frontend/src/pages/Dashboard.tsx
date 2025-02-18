import { useState, useEffect, SetStateAction } from 'react'
import {
  Box,
  Button,
  FormControl,
  Input,
  VStack,
  Heading,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
  HStack,
  Text,
  useToast,
} from '@chakra-ui/react'
import { CopyIcon, DeleteIcon, ExternalLinkIcon } from '@chakra-ui/icons'
import axios from 'axios'

interface UrlEntry {
  id: string
  originalUrl: string
  shortCode: string
  createdAt: string
  clicks: number
}

const Dashboard = () => {
  const [url, setUrl] = useState('')
  const [urls, setUrls] = useState<UrlEntry[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const toast = useToast()

  const loadUrls = async () => {
    try {
      const response = await axios.get('/api/shorten')
      setUrls(response.data)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load URLs',
        status: 'error',
        duration: 3000,
      })
    }
  }

  useEffect(() => {
    loadUrls()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      await axios.post('/api/shorten', { url })
      setUrl('')
      loadUrls()
      toast({
        title: 'Success',
        description: 'URL shortened successfully',
        status: 'success',
        duration: 3000,
      })
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to shorten URL',
        status: 'error',
        duration: 3000,
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleDelete = async (shortCode: string) => {
    try {
      await axios.delete(`/api/shorten/${shortCode}`)
      loadUrls()
      toast({
        title: 'Success',
        description: 'URL deleted successfully',
        status: 'success',
        duration: 3000,
      })
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete URL',
        status: 'error',
        duration: 3000,
      })
    }
  }

  const handleCopy = (shortCode: string) => {
    const shortUrl = `${window.location.origin}/${shortCode}`
    navigator.clipboard.writeText(shortUrl)
    toast({
      title: 'Success',
      description: 'URL copied to clipboard',
      status: 'success',
      duration: 2000,
    })
  }

  return (
    <Box maxW="container.lg" mx="auto">
      <VStack spacing={8} align="stretch">
        <Box>
          <Heading size="lg" mb={4}>Shorten URL</Heading>
          <form onSubmit={handleSubmit}>
            <HStack>
              <FormControl isRequired>
                <Input
                  type="url"
                  placeholder="Enter URL to shorten"
                  value={url}
                  onChange={(e: { target: { value: SetStateAction<string> } }) => setUrl(e.target.value)}
                />
              </FormControl>
              <Button
                type="submit"
                colorScheme="blue"
                isLoading={isLoading}
                minW="150px"
              >
                Shorten
              </Button>
            </HStack>
          </form>
        </Box>

        <Box>
          <Heading size="lg" mb={4}>Your URLs</Heading>
          {urls.length === 0 ? (
            <Text>No URLs shortened yet</Text>
          ) : (
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>Original URL</Th>
                  <Th>Short URL</Th>
                  <Th>Clicks</Th>
                  <Th>Created</Th>
                  <Th>Actions</Th>
                </Tr>
              </Thead>
              <Tbody>
                {urls.map((entry) => (
                  <Tr key={entry.id}>
                    <Td maxW="300px" isTruncated>{entry.originalUrl}</Td>
                    <Td>
                      <HStack>
                        <Text>{entry.shortCode}</Text>
                        <IconButton
                          aria-label="Copy URL"
                          icon={<CopyIcon />}
                          size="sm"
                          onClick={() => handleCopy(entry.shortCode)}
                        />
                        <IconButton
                          aria-label="Open URL"
                          icon={<ExternalLinkIcon />}
                          size="sm"
                          as="a"
                          href={`/${entry.shortCode}`}
                          target="_blank"
                        />
                      </HStack>
                    </Td>
                    <Td>{entry.clicks}</Td>
                    <Td>{new Date(entry.createdAt).toLocaleDateString()}</Td>
                    <Td>
                      <IconButton
                        aria-label="Delete URL"
                        icon={<DeleteIcon />}
                        colorScheme="red"
                        size="sm"
                        onClick={() => handleDelete(entry.shortCode)}
                      />
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          )}
        </Box>
      </VStack>
    </Box>
  )
}

export default Dashboard 