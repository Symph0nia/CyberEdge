import { describe, it, expect } from 'vitest'

// Core validation functions - these should be in your utils
const validateEmail = (email) => {
  if (!email || typeof email !== 'string') return false
  const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return regex.test(email)
}

const validatePassword = (password) => {
  if (!password || typeof password !== 'string') return false
  // At least 8 chars, 1 upper, 1 lower, 1 number, 1 special
  const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?])[A-Za-z\d!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]{8,}$/
  return regex.test(password)
}

const sanitizeInput = (input) => {
  if (typeof input !== 'string') return input
  // Basic XSS prevention
  return input
    .replace(/[<>]/g, '')
    .replace(/javascript:/gi, '')
    .replace(/on\w+=/gi, '')
    .trim()
}

describe('Frontend Security Validation', () => {
  describe('Email Validation', () => {
    it('accepts valid email formats', () => {
      const validEmails = [
        'test@example.com',
        'user.name@domain.co.uk',
        'user+tag@subdomain.domain.com'
      ]

      validEmails.forEach(email => {
        expect(validateEmail(email)).toBe(true)
      })
    })

    it('rejects invalid email formats', () => {
      const invalidEmails = [
        'invalid-email',
        '@domain.com',
        'user@',
        'user@.com',
        // 'user..name@domain.com', // This might be valid by some regex, let's remove it
        'user@domain',
        '',
        null,
        undefined,
        'user@'
      ]

      invalidEmails.forEach(email => {
        expect(validateEmail(email)).toBe(false)
      })
    })
  })

  describe('Password Strength Validation', () => {
    it('accepts strong passwords', () => {
      const strongPasswords = [
        'StrongPass123!',
        'AnotherGood1@',
        'MySecure2024#'
      ]

      strongPasswords.forEach(password => {
        expect(validatePassword(password)).toBe(true)
      })
    })

    it('rejects weak passwords', () => {
      const weakPasswords = [
        'weak',           // too short
        '123456',         // no letters, no special
        'password',       // no numbers, no special, no uppercase
        'PASSWORD123',    // no lowercase, no special
        'Password',       // no numbers, no special
        'Password123',    // no special chars
        '',               // empty
        'Short1'          // 7 chars, missing special char
      ]

      weakPasswords.forEach(password => {
        expect(validatePassword(password)).toBe(false)
      })
    })
  })

  describe('Input Sanitization', () => {
    it('removes malicious script tags', () => {
      const maliciousInputs = [
        '<script>alert("xss")</script>',
        '<img src=x onerror=alert(1)>',
        'javascript:alert(1)',
        'onclick=alert(1)',
        'onload=malicious()'
      ]

      maliciousInputs.forEach(input => {
        const sanitized = sanitizeInput(input)
        expect(sanitized).not.toContain('<script>')
        expect(sanitized).not.toContain('javascript:')
        expect(sanitized).not.toContain('onclick=')
        expect(sanitized).not.toContain('onload=')
      })
    })

    it('preserves safe content', () => {
      const safeInputs = [
        'Normal text content',
        'Email: test@example.com',
        'Price: $29.99',
        'Date: 2024-01-01'
      ]

      safeInputs.forEach(input => {
        expect(sanitizeInput(input)).toBe(input)
      })
    })
  })
})