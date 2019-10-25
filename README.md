[![Go Report Card](https://goreportcard.com/badge/github.com/ilyapt/fabric-certstore)](https://goreportcard.com/report/github.com/ilyapt/fabric-certstore) [![Build Status](https://travis-ci.com/ilyapt/fabric-certstore.svg?branch=master)](https://travis-ci.com/ilyapt/fabric-certstore)

# Уменьшение размера транзакции Hyperledger Fabric
Основано на задаче [FAB-8007](https://jira.hyperledger.org/browse/FAB-8007) by Yacov Manevich, IBM Research, Hyperledger Fabric Maintainer

Спроектировано совместно с Artem Barger, IBM Research, Hyperledger Fabric Maintainer

Моя презентация по этому проекту с Hyperledger Bootcamp Moscow доступна [здесь](https://docs.google.com/presentation/d/1gEUovzmY9p_Nca88L_xwohv0g4lgTFyBRiWx1zCBghw/edit?usp=sharing)

Дисклеймер: *данный код не претендует на production-ready, это скорее исследование возможности уменьшения размера транзакции в Hyperledger Fabric и мы продолжаем тестировать работоспособность этого кода в наших проектах. Используя этот код вы принимаете ответственность на себя и освобождаете нас от всех и любых претензий, исков, убытков, обязательств, ущерба и расходов (включая судебные издержки), возникших в связи использовании этого кода.*

- Проект делается по принципу наименьшего вмешательства в оригинальный код fabric и fabric-sdk-go, чтобы его было проще поддерживать и переходить на новые версии оригинального кода.
- Есть ряд открытых проблем, одной из наиболее значимых является отсутствие сертификата endorser в ответах при query запросах, соответственно sdk не может убедиться, что получило ответ от настоящего пира.