import iconv from 'iconv-lite'
import { Buffer } from 'buffer'

export type DetectedEncoding = 'utf-8' | 'utf-8-bom' | 'utf-16le' | 'utf-16be' | 'gb18030' | 'big5'
export type OutputEncoding = 'keep' | 'utf-8' | 'utf-8-bom'

export interface DecodedText {
  text: string
  encoding: DetectedEncoding
  hadBOM: boolean
}

function replacementRatio(text: string): number {
  if (text.length === 0) return 1
  let count = 0
  for (const char of text) if (char === '\ufffd') count++
  return count / [...text].length
}

function decodeUtf8(data: Uint8Array, fatal = false): string {
  return new TextDecoder('utf-8', { fatal }).decode(data)
}

export function decodeText(data: Uint8Array): DecodedText {
  if (data.length >= 3 && data[0] === 0xef && data[1] === 0xbb && data[2] === 0xbf) {
    return { text: decodeUtf8(data.subarray(3)), encoding: 'utf-8-bom', hadBOM: true }
  }
  if (data.length >= 2 && data[0] === 0xff && data[1] === 0xfe) {
    return {
      text: new TextDecoder('utf-16le').decode(data.subarray(2)),
      encoding: 'utf-16le',
      hadBOM: true,
    }
  }
  if (data.length >= 2 && data[0] === 0xfe && data[1] === 0xff) {
    // TextDecoder supports utf-16be in current Chrome, Edge, Firefox and Safari.
    return {
      text: new TextDecoder('utf-16be').decode(data.subarray(2)),
      encoding: 'utf-16be',
      hadBOM: true,
    }
  }

  try {
    return { text: decodeUtf8(data, true), encoding: 'utf-8', hadBOM: false }
  } catch {
    const input = Buffer.from(data)
    const gbText = iconv.decode(input, 'gb18030')
    const big5Text = iconv.decode(input, 'big5')
    const gbBad = replacementRatio(gbText)
    const big5Bad = replacementRatio(big5Text)
    if (gbBad <= 0.02 || gbBad <= big5Bad) {
      return { text: gbText, encoding: 'gb18030', hadBOM: false }
    }
    return { text: big5Text, encoding: 'big5', hadBOM: false }
  }
}

export function encodeText(
  text: string,
  output: OutputEncoding,
  sourceEncoding: DetectedEncoding,
): Uint8Array {
  if (output === 'utf-8-bom') {
    const body = new TextEncoder().encode(text)
    return new Uint8Array([0xef, 0xbb, 0xbf, ...body])
  }
  if (output === 'utf-8') return new TextEncoder().encode(text)

  if (sourceEncoding === 'gb18030' || sourceEncoding === 'big5') {
    return new Uint8Array(iconv.encode(text, sourceEncoding))
  }
  // UTF-16 input and ordinary UTF-8 input are intentionally emitted as UTF-8,
  // matching the desktop encoder's safe fallback for unsupported keep targets.
  return new TextEncoder().encode(text)
}

export async function readFileText(file: File): Promise<DecodedText> {
  return decodeText(new Uint8Array(await file.arrayBuffer()))
}
