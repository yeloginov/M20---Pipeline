// Программа реализует пайплайн:
//   1. Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
//   2. Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0.
//   3. Стадия буферизации. Накапливаем полученные данные в буфер. Буфер опустошается через
//	    определенные промежутки времени.
// Написал Егор Логинов (GO-10) по заданию SkillFactory в модуле 20

package main

import (
	"container/ring"
	"fmt"
	"strconv"
	"time"
)

// Размер буфера при обработке в пайплайне
const buffSize int = 5

// Таймаут передачи данных из буфера в секундах
const buffTime int = 10

// Инициализирующее значение для кольца
const rNil int = -1

// Метод над кольцом ring - подсчет элементов НЕrNil
func rNum(r *ring.Ring) int {
	count := 0
	for i := 0; i < r.Len(); i++ {
		if r.Value != rNil {
			count++
		}
		r = r.Next()
	}

	return count
}

// Функция getInt() первым значением возвращает введенное с консоли целое число.
// Дополнительно передает исходную введенную строку для возможного парсинга на стороне вызова
// и ошибку.
func getInt() (int, string, error) {
	var s string

	_, err := fmt.Scanln(&s)
	if err != nil {
		return -1, s, err
	}

	i, err := strconv.ParseInt(s, 0, 0)
	if err != nil {
		return -1, s, err
	}

	return int(i), s, nil
}

func main() {
	// Источник данных для конвейера:
	// Горутина ожидает ввода данных в консоль и передает их в канал
	// обработки пайплайна. Вторым значением возвращает канал синхронизации завершения
	srcstream := func() (<-chan int, <-chan int) {
		// Возвращаемый канал с данными для пайплайна
		output := make(chan int)
		// Канал синхронизации завершения работы горутин
		done := make(chan int)
		go func() {
			for {
				// Получаем число из консоли
				fmt.Println("Введите число (или 'exit' для выхода): ")
				n, s, err := getInt()

				// Обрабатываем ввод команды 'exit' - закрываем канал завершения
				if s == "exit" {
					close(done)
					return
				}

				// Обрабатываем некорректный ввод
				if err != nil {
					fmt.Printf("Некорректный ввод, %v не является числом, попробуйте еще раз.\n", s)
					continue
				}

				// Если введено число, передаем его в канал
				output <- n
			}
		}()
		return output, done

	}

	// Потребитель данных, полученных из пайплайна
	receiver := func(done <-chan int, instream <-chan int) {
		go func() {
			for {
				select {
				// Слушаем канал входного потока
				case n := <-instream:
					fmt.Printf("[log] Потребитель получил на вход значение %v, спасибо! Пошел с ним работать, жду еще...\n", n)
				// Завершаем горутину при сигнале из канала done
				case <-done:
					return
				}
			}
		}()
	}

	// Замыкание, фильтрующее отрицательные значения в пайплайне
	filter1 := func(done <-chan int, input <-chan int) <-chan int {
		// Возвращаемый канал с данными для пайплайна
		output := make(chan int)
		go func() {
			for {
				select {
				// Слушаем канал входного потока и передаем на выход только
				// не отрицательные значения
				case n := <-input:
					if n < 0 {
						fmt.Println("[log] Хм.., отрицательные не пропускаем...")
					} else {
						output <- n
					}
					// Завершаем горутину при сигнале из канала done
				case <-done:
					return
				}
			}
		}()
		return output
	}

	// Замыкание, фильтрующее значения не кратные 3
	filter2 := func(done <-chan int, input <-chan int) <-chan int {
		// Возвращаемый канал с данными для пайплайна
		output := make(chan int)
		go func() {
			for {
				select {
				// Слушаем канал входного потока и передаем на выход только
				// значения кратные 3
				case n := <-input:
					if n == 0 || n%3 != 0 {
						fmt.Printf("[log] Хм.., число %v не кратно 3, не пропускаем...\n", n)
					} else {
						output <- n
					}
					// Завершаем горутину при сигнале из канала done
				case <-done:
					return
				}
			}
		}()
		return output
	}

	// Замыкание буферизации. Накапливаем полученные данные в буфер. Буфер опустошается через
	// определенные промежутки времени.
	buffering := func(done <-chan int, input <-chan int) <-chan int {
		// Возвращаемый канал с данными для пайплайна
		output := make(chan int)

		// Создадим кольцо под буффер, инициируем начальными значениями
		buff := ring.New(buffSize)
		for i := 0; i < buff.Len(); i++ {
			buff.Value = rNil
			buff = buff.Next()
		}

		// Первая горутина будет просто собирать данные в буфер
		go func() {
			for {
				select {
				case n := <-input:
					buff.Value = n
					buff = buff.Next()
				case <-done:
					return
				}
			}
		}()

		// Вторая горутина будет сбрасывать данные из буфера по таймеру
		go func() {
			for {
				// Создаем таймер
				t := time.NewTimer(time.Second * time.Duration(buffTime))
				select {
				// При срабатывании таймера, сбрасываем данные из буфера в канал вывода
				case <-t.C:
					if rNum(buff) == 0 {
						fmt.Printf("[log] За прошедшие %v сек. новых данных в буфер не поступило, ждем...\n", buffTime)
					} else {
						for i := 0; i < buff.Len(); i++ {
							if buff.Value != rNil {
								output <- buff.Value.(int)
								buff.Value = rNil
							}
							buff = buff.Next()
						}
					}
				case <-done:
					return
				}
			}
		}()

		return output
	}

	// Инициализируем канал входных данных и закрывающий канал
	data, complete := srcstream()
	// Передаем данные полтребителю, обрабатывая их в пайплайне
	receiver(complete, buffering(complete, filter2(complete, filter1(complete, data))))

	// Блокируем основную горутину для ожидания завершения
	select {
	case <-complete:
		return
	}

}
