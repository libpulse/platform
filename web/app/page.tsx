'use client'

import { useEffect, useState } from 'react'
import type { Session, User } from '@supabase/supabase-js'
import { supabaseClient } from '@/lib/supabaseClient'

type AuthState = {
  loading: boolean
  session: Session | null
  user: User | null
  error: string | null
}

export default function HomePage() {
  const [auth, setAuth] = useState<AuthState>({
    loading: true,
    session: null,
    user: null,
    error: null,
  })

  useEffect(() => {
    const getInitialSession = async () => {
      const { data, error } = await supabaseClient.auth.getSession()

      setAuth({
        loading: false,
        session: data.session,
        user: data.session?.user ?? null,
        error: error ? error.message : null,
      })
    }

    getInitialSession()

    const {
      data: { subscription },
    } = supabaseClient.auth.onAuthStateChange((_event, session) => {
      setAuth((prev) => ({
        ...prev,
        session: session,
        user: session?.user ?? null,
        error: null,
      }))
    })

    return () => {
      subscription.unsubscribe()
    }
  }, [])

  // Check User
  useEffect(() => {
    const checkUser = async () => {
      const { data, error } = await supabaseClient.auth.getUser();
      console.log("USER:", data?.user, "ERROR:", error);
    };
    checkUser();
  }, []);


  const handleLoginWithGithub = async () => {
    setAuth((prev) => ({ ...prev, loading: true, error: null }))

    const { error } = await supabaseClient.auth.signInWithOAuth({
      provider: 'github',
      options: {
        redirectTo: `${window.location.origin}`,
      },
    })

    if (error) {
      setAuth((prev) => ({
        ...prev,
        loading: false,
        error: error.message,
      }))
    }
  }

  const handleLogout = async () => {
    setAuth((prev) => ({ ...prev, loading: true, error: null }))
    const { error } = await supabaseClient.auth.signOut()

    setAuth((prev) => ({
      ...prev,
      loading: false,
      session: null,
      user: null,
      error: error ? error.message : null,
    }))
  }

  const { loading, user, error } = auth

  return (
    <main className="min-h-screen flex flex-col items-center justify-center gap-6">
      <h1 className="text-2xl font-semibold">LibPulse – Login (GitHub via Supabase)</h1>

      {loading && <p>Checking session...</p>}

      {!loading && !user && (
        <>
          <p>You are not logged in.</p>
          <button
            onClick={handleLoginWithGithub}
            className="px-4 py-2 rounded-md border text-sm"
          >
            Sign in with GitHub
          </button>
        </>
      )}

      {!loading && user && (
        <div className="flex flex-col items-center gap-3">
          <p className="text-sm">
            Logged in as <span className="font-mono">{user.email ?? user.id}</span>
          </p>
          <button
            onClick={handleLogout}
            className="px-4 py-2 rounded-md border text-sm"
          >
            Sign out
          </button>
        </div>
      )}

      {error && <p className="text-red-500 text-sm">Error: {error}</p>}
    </main>
  )
}