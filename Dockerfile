# Написал Егор Логинов (GO-11) по итоговому заданию SkillFactory в модуле 26
# v.1.1 в задании из модуля 26a.3

FROM golang AS pipeline_build
RUN mkdir -p /go/src/pipeline
# Устанавливаем рабочую директорию для последующих инструкций, где она будет как "." 
WORKDIR /go/src/pipeline
# Добавляем файлы main.go и go.mod в папку /go/src/pipeline
ADD main.go .
ADD go.mod .
# Компилируем программы и помещаем запускаемый файл в /go/bin/
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Yegor Loginov<loginov@gmail.com>"
LABEL description="Docker Container for SkillFactory task 26.3"
#RUN mkdir -p /app
WORKDIR /app
# Копируем запускаемый файл из предыдущего контенера в папку /app
COPY --from=pipeline_build /go/bin/pipeline .
ENTRYPOINT ./pipeline

# Memo работы в терминале:
#
# Построение образа:
# docker build -t sf_pipeline .
#
# Запуск контейнера:
# docker run -d --name=pipeline --rm sf_pipeline
#
# Вход в консоль контейнера (в образе alpine нет bash, вызываем sh):
# docker exec -it pipeline /bin/sh
#
# Выход из консоли контейнера
# exit
# 
# Остановка контейнера (при этом, благодаря инструкции --rm, он удалится)
# docker stop pipeline