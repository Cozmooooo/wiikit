import { writable } from 'svelte/store'

export type Game = {
  path: string
  title: string
  gameId: string
  region: string
  sizeBytes: number
  selected?: boolean
}

export type DriveInfo = {
  path: string
  label: string
  sizeBytes: number
  isRemovable: boolean
}

export type Progress = {
  gameId?: string
  title?: string
  bytesDone?: number
  bytesTotal?: number
  stage?: string
}

export type ConsoleType = 'wii' | 'wiiu' | 'wiimini'

export const step = writable<0 | 1 | 2 | 3>(0)
export const consoleType = writable<ConsoleType>('wii')
export const library = writable<Game[]>([])
export const drives = writable<DriveInfo[]>([])
export const usbDrive = writable<DriveInfo | null>(null)
export const sdDrive = writable<DriveInfo | null>(null)
export const installing = writable(false)
export const log = writable<string[]>([])
export const progress = writable<Progress>({})
export const formatConfirmed = writable(false)
export const wiilinkConfirmed = writable(false)
