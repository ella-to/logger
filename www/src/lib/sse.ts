import { useEffect, useState, useDeferredValue, useMemo } from 'react'

export function useSSE<T>(addr: string): [T[], () => void] {
  const [messages, setMessages] = useState<T[]>([])

  useEffect(() => {
    const eventSource = new EventSource(addr, { withCredentials: true })

    eventSource.onmessage = event => {
      setMessages(prevMessages => [...prevMessages, JSON.parse(event.data)])
    }

    return () => {
      eventSource.close()
    }
  }, [addr])

  return [messages, () => setMessages([])]
}

export function useProcess<T, V>(arr: T[], fn: (arg: T[]) => V[]): V[] {
  const deferredArr = useDeferredValue(arr)

  const processed = useMemo(() => {
    return fn(deferredArr)
  }, [deferredArr])

  return processed
}
