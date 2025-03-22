package utils

func IsValidLuhn(number string) bool {
    sum := 0
    parity := len(number) % 2
    for i, char := range number {
        digit := int(char - '0')
        if i%2 == parity {
            digit *= 2
            if digit > 9 {
                digit -= 9
            }
        }
        sum += digit
    }
    return sum%10 == 0
}