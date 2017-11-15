FROM funcy/go
WORKDIR /fnproject
ADD fn-mock-vista /fnproject/fn-mock-vista
CMD ["/fnproject/fn-mock-vista"]
