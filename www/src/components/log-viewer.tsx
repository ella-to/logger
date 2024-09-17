'use client'

import React, { useState } from 'react'
import { ChevronRight, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table'
import { useProcess, useSSE } from '@/lib/sse'

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

type LogEntry = {
  id: string
  message: string
  level: LogLevel
  timestamp: string
  meta: {
    aggregate_id?: string
    pkg?: string
    fn?: string
    [key: string]: string | number | undefined
  }
  children?: LogEntry[]
}

const levelColors: Record<LogLevel, string> = {
  DEBUG: 'bg-purple-200 hover:bg-purple-300',
  INFO: 'bg-green-200 hover:bg-green-300',
  WARN: 'bg-yellow-200 hover:bg-yellow-300',
  ERROR: 'bg-red-200 hover:bg-red-300',
}

const LogLegend: React.FC = () => (
  <div className="flex space-x-4 mb-4">
    {Object.entries(levelColors).map(([level, colorClass]) => (
      <div key={level} className="flex items-center">
        <div className={`w-4 h-4 rounded ${colorClass.split(' ')[0]} mr-2`} />
        <span className="text-sm capitalize">{level}</span>
      </div>
    ))}
  </div>
)

const LogEntryRow: React.FC<{
  entry: LogEntry
  depth: number
  onSelect: (entry: LogEntry) => void
  isSelected: boolean
}> = ({ entry, depth, onSelect, isSelected }) => {
  const [isExpanded, setIsExpanded] = useState(depth < 2)

  const toggleExpand = (e: React.MouseEvent) => {
    e.stopPropagation()
    setIsExpanded(!isExpanded)
  }

  return (
    <>
      <TableRow
        className={`${levelColors[entry.level]} transition-colors cursor-pointer`}
        onClick={() => onSelect(entry)}>
        <TableCell className="w-full">
          <div className="flex items-center space-x-2">
            <div style={{ width: `${depth * 20}px` }} className="flex-shrink-0" />
            {entry.children && entry.children.length > 0 && (
              <Button variant="ghost" size="sm" className="p-0 h-6 w-6" onClick={toggleExpand}>
                {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              </Button>
            )}
            {(!entry.children || entry.children.length === 0) && <div className="w-6" />}
            <div className="flex flex-col">
              <span className="text-sm font-medium">{entry.message}</span>
              <span className="text-xs text-muted-foreground">
                {entry.meta.pkg && entry.meta.fn && `${entry.meta.pkg}.${entry.meta.fn}`}
              </span>
            </div>
          </div>
        </TableCell>
      </TableRow>
      {isExpanded &&
        entry.children &&
        entry.children.map(child => (
          <LogEntryRow key={child.id} entry={child} depth={depth + 1} onSelect={onSelect} isSelected={isSelected} />
        ))}
    </>
  )
}

const LogDetails: React.FC<{ selectedLog: LogEntry | null }> = ({ selectedLog }) => {
  if (!selectedLog) {
    return <div className="p-4 text-center text-muted-foreground">Select a log entry to view details</div>
  }

  return (
    <div className="p-4 space-y-4">
      <h3 className="text-lg font-semibold">{selectedLog.message}</h3>
      <div>
        <div>
          <span className="font-medium">Level:</span> {selectedLog.level}
        </div>
        <div>
          <span className="font-medium">Timestamp:</span> {selectedLog.timestamp.toLocaleString()}
        </div>
        {Object.entries(selectedLog.meta).map(([key, value]) => (
          <div key={key} className="col-span-2">
            <span className="font-medium">{key}:</span> {value}
          </div>
        ))}
      </div>
    </div>
  )
}

const LogTable: React.FC<{
  logs: LogEntry[]
  selectedLog: LogEntry | null
  onSelectLog: (log: LogEntry) => void
}> = ({ logs, selectedLog, onSelectLog }) => (
  <Table>
    <TableBody>
      {logs.map(log => (
        <LogEntryRow
          key={log.id}
          entry={log}
          depth={0}
          onSelect={onSelectLog}
          isSelected={selectedLog?.id === log.id}
        />
      ))}
    </TableBody>
  </Table>
)

function sortLogEntriesByAggregateId(logs: LogEntry[]): LogEntry[][] {
  var bucketsIdIdx: Record<string, number> = {}
  var buckets: LogEntry[][] = []
  logs.forEach(log => {
    const aggregateId = log.meta.aggregate_id

    if (!aggregateId) {
      buckets.push([log])
      return
    }

    let idx = bucketsIdIdx[aggregateId]
    let bucket: LogEntry[]
    if (idx === undefined) {
      idx = buckets.length
      bucketsIdIdx[aggregateId] = idx
      bucket = []
      buckets.push(bucket)
    } else {
      bucket = buckets[idx]
    }

    bucket.push(log)
  })

  return buckets
}

function sortBucketByTimestamp(bucket: LogEntry[]): LogEntry[] {
  return bucket.sort((a, b) => {
    return new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  })
}

function build(logs: LogEntry[]): LogEntry {
  const root = logs[0]

  if (logs.length === 1) {
    return root
  }

  root.children = logs.slice(1)

  return root
}

function process(logs: LogEntry[]): LogEntry[] {
  const buckets = sortLogEntriesByAggregateId(sortBucketByTimestamp(logs))

  return buckets.map(bucket => {
    return build(bucket)
  })
}

export function LogViewerComponent() {
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)
  const rawLogs = useSSE<LogEntry>('/logs')
  const logs = useProcess(rawLogs, process)

  console.log(logs)

  return (
    <div className="w-full h-screen flex flex-col bg-background">
      <div className="p-4 border-b">
        <h2 className="text-2xl font-bold mb-4">Log Viewer</h2>
        <LogLegend />
      </div>
      <div className="flex-grow flex flex-col overflow-hidden">
        <div className="flex-grow overflow-hidden">
          <ScrollArea className="h-full w-full">
            <LogTable logs={logs} selectedLog={selectedLog} onSelectLog={setSelectedLog} />
          </ScrollArea>
        </div>
        <div className="h-1/3 border-t">
          <ScrollArea className="h-full w-full">
            <LogDetails selectedLog={selectedLog} />
          </ScrollArea>
        </div>
      </div>
    </div>
  )
}
