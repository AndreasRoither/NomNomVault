import { createFileRoute } from '@tanstack/solid-router'
import { HomePage } from '../components/HomePage'

export const Route = createFileRoute('/')({ component: HomePage })
